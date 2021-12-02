package overseer

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"git.qutoutiao.net/gopher/qms/third_party/forked/jpillora/overseer/restart"
)

var tmpBinPath = filepath.Join(os.TempDir(), "overseer-"+token())

//a overseer master process
type master struct {
	*Config
	logger
	slaveID             int
	slaveCmd            atomic.Value
	tempSlaveCmd        *exec.Cmd // slave ready前,临时保存
	slaveExtraFiles     []*os.File
	binPath, tmpBinPath string
	binPerms            os.FileMode
	binHash             []byte
	restart             *restart.Restart
	restartingC         chan struct{}
	printCheckUpdate    bool
}

func newMaster(cfg *Config) *master {
	mp := &master{
		Config:      cfg,
		logger:      newWrapLogger("[overseer master] ", cfg.Logger),
		restartingC: make(chan struct{}, 1),
	}
	mp.restart = restart.New(mp, cfg.RestartTimeout)

	return mp
}

func (mp *master) run() error {
	mp.logger.Debugf("run")
	if err := mp.checkBinary(); err != nil {
		return err
	}
	if mp.Config.Fetcher != nil {
		if err := mp.Config.Fetcher.Init(); err != nil {
			mp.logger.Warnf("fetcher init failed (%s). fetcher disabled.", err)
			mp.Config.Fetcher = nil
		}
	}
	//updater-forker comms
	mp.setupRestartListen()
	if err := mp.retreiveFileDescriptors(); err != nil {
		return err
	}
	go mp.restart.Run() // first
	mp.setupSignalling()
	//if mp.Config.Fetcher != nil {
	//	mp.printCheckUpdate = true
	//	mp.fetch()
	//	go mp.fetchLoop()
	//}
	return nil
}

// restart begin
func (mp *master) RestartBegin() error {
	mp.logger.Debugf("restart begin")
	return nil
}

// restart start[fork new slave and run]
func (mp *master) ProgramStart() error {
	mp.logger.Debugf("starting %s", mp.binPath)
	mp.slaveID++
	cmd := exec.Command(mp.binPath)
	mp.tempSlaveCmd = cmd
	e := os.Environ()
	e = append(e, envMasterVersion+"="+masterVersion)
	e = append(e, envAddresses+"="+strings.Join(mp.Addresses, "-"))
	e = append(e, envRestartPort+"="+strconv.Itoa(mp.RestartPort))
	e = append(e, envBinID+"="+hex.EncodeToString(mp.binHash))
	e = append(e, envBinPath+"="+mp.binPath)
	e = append(e, envSlaveID+"="+strconv.Itoa(mp.slaveID))
	e = append(e, envIsSlave+"=1")
	e = append(e, envNumFDs+"="+strconv.Itoa(len(mp.slaveExtraFiles)))

	cmd.Env = e
	//inherit master args/stdfiles
	cmd.Args = os.Args
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//include socket files
	cmd.ExtraFiles = mp.slaveExtraFiles
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start slave process: %s", err)
	}

	go func() {
		//convert wait into channel
		cmdwait := make(chan error, 1)
		go func() {
			cmdwait <- cmd.Wait()
		}()

		//wait....
		select {
		case err := <-cmdwait:
			if !mp.isCurrentSlave(cmd) {
				mp.restart.ProgramFailed(err)
				return
			}

			// 只有主slave异常退出 才会让服务退出
			var code int
			if err != nil {
				code = 1
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						code = status.ExitStatus()
					}
				}
			}
			qlog.WithError(err).Infof("prog exited with %d", code)
			os.Exit(code)
		case <-mp.restartingC: // 说明正在reload, 老的slave主动让位
		}
	}()

	return nil
}

// restart ready[新老slave替换]
func (mp *master) ProgramReady() (*exec.Cmd, error) {
	old, _ := mp.slaveCmd.Load().(*exec.Cmd)
	mp.slaveCmd.Store(mp.tempSlaveCmd)
	mp.tempSlaveCmd = nil
	mp.logger.Debugf("slave[%d] ready.", mp.slaveID)

	return old, nil
}

// restart failed[新slave run error]
func (mp *master) ProgramFailed(e error) error {
	// 打印error
	mp.logger.Errorf("slave[%d] restart error: %v", mp.slaveID, e)
	if e == restart.ErrTimeout {
		slave := mp.tempSlaveCmd
		mp.tempSlaveCmd = nil
		if mp.isCurrentSlave(slave) { // 如果是主进程直接退出
			os.Exit(1)
		}

		if err := slave.Process.Signal(os.Kill); err != nil {
			os.Exit(1)
		}
	}

	// 通知此次reload失败
	return e
}

// restart end[关闭老slave]
func (mp *master) RestartEnd(old *exec.Cmd) error {
	if old != nil {
		mp.restartingC <- struct{}{} // 因为第一次启动时没有old 所以不应该释放restartingC信号
		mp.logger.Debugf("send signal to old process[%s]", mp.RestartSignal)
		if err := old.Process.Signal(mp.RestartSignal); err != nil {
			mp.logger.Debugf("signal failed (%s), assuming slave process died unexpectedly", err)
			os.Exit(1)
		}

		go func(slaveId int) {
			time.Sleep(mp.Config.TerminateTimeout)
			// [进程退出]或[cmd执行Run/Wait阻塞方法返回后]才有值
			if old.ProcessState == nil {
				mp.logger.Debugf("graceful timeout, forcing slave#%d exit", slaveId)
				old.Process.Kill()
			}
		}(mp.slaveID - 1)
	}
	return nil
}

func (mp *master) isCurrentSlave(slave *exec.Cmd) bool {
	current, _ := mp.slaveCmd.Load().(*exec.Cmd)
	if current == nil { // 说明是第一次启动
		return true
	}

	return current == slave
}

func (mp *master) checkBinary() error {
	//get path to binary and confirm its writable
	//binPath, err := osext.Executable()  //这种方式在linux环境下，自动转换了软连接
	binPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		return fmt.Errorf("failed to find binary path (%s)", err)
	}
	mp.logger.Debugf("binPath=%s", binPath)
	mp.binPath = binPath
	if info, err := os.Stat(binPath); err != nil {
		return fmt.Errorf("failed to stat binary (%s)", err)
	} else if info.Size() == 0 {
		return fmt.Errorf("binary file is empty")
	} else {
		//copy permissions
		mp.binPerms = info.Mode()
	}
	f, err := os.Open(binPath)
	if err != nil {
		return fmt.Errorf("cannot read binary (%s)", err)
	}
	//initial hash of file
	hash := sha1.New()
	io.Copy(hash, f)
	mp.binHash = hash.Sum(nil)
	f.Close()
	//test bin<->tmpbin moves
	if mp.Config.Fetcher != nil {
		if err := move(tmpBinPath, mp.binPath); err != nil {
			return fmt.Errorf("cannot move binary (%s)", err)
		}
		if err := move(mp.binPath, tmpBinPath); err != nil {
			return fmt.Errorf("cannot move binary back (%s)", err)
		}
	}
	return nil
}

func (mp *master) setupSignalling() {
	//read all master process signals
	signals := make(chan os.Signal)
	signal.Notify(signals)
	for s := range signals {
		mp.handleSignal(s)
	}
}

func (mp *master) handleSignal(s os.Signal) {
	if s == mp.RestartSignal { // restart step one
		mp.logger.Infof("receive reload signal[%s]", mp.RestartSignal)
		go mp.restart.Run()
	} else if s.String() == "child exited" {
		// will occur on every restart, ignore it
	} else if mp.tempSlaveCmd != nil && s == SIGUSR1 {
		mp.logger.Tracef("receive SIGUSR1")
		go mp.restart.ProgramReady()
	} else if slave, _ := mp.slaveCmd.Load().(*exec.Cmd); slave != nil && slave.Process != nil {
		//while the slave process is running, proxy all signals through
		mp.logger.Tracef("proxy signal (%s)", s)
		if err := slave.Process.Signal(s); err != nil {
			mp.logger.Errorf("signal failed (%s), assuming slave process died unexpectedly", err)
			os.Exit(1)
		}
	} else if s == os.Interrupt { //otherwise if not running, kill on CTRL+c
		mp.logger.Debugf("interupt with no slave")
		os.Exit(1)
	} else {
		mp.logger.Tracef("signal discarded (%s), no slave process", s)
	}
}

func (mp *master) setupRestartListen() {
	addr := fmt.Sprintf("127.0.0.1:%d", mp.RestartPort)
	mux := http.NewServeMux()
	mux.HandleFunc("/reload", func(writer http.ResponseWriter, request *http.Request) {
		if token, ok := request.URL.Query()["token"]; ok && token[0] == "smq" {
			mp.logger.Infof("Got /reload from %s", addr)
			//user initiated manual restart
			if err := mp.restart.Run(); err != nil {
				qlog.Errorf("restart failed: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(fmt.Sprintf("restart failed: %v", err)))
				return
			}
			writer.WriteHeader(http.StatusOK)
		} else {
			mp.logger.Warnf("Got /reload from %s, but invalid token", addr)
			writer.WriteHeader(http.StatusUnauthorized)
		}
	})
	s := http.Server{
		Addr:    addr,
		Handler: mux,
	}
	if mp.checkPortUsed(addr) {
		mp.logger.Warnf("listen address(%s) has in used", addr)
		os.Exit(1)
	}
	go func() {
		mp.logger.Debugf("ListenAndServe(%s)", addr)
		err := s.ListenAndServe()
		if err != nil {
			mp.logger.Warnf("graceful reload loop exit: %s", err.Error())
		}
	}()
}

func (mp *master) checkPortUsed(addr string) bool {
	conn, _ := net.DialTimeout("tcp", addr, time.Second)
	if conn != nil {
		conn.Close()
		return true
	}
	return false
}

func (mp *master) retreiveFileDescriptors() error {
	mp.slaveExtraFiles = make([]*os.File, len(mp.Config.Addresses))
	for i, addr := range mp.Config.Addresses {
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return fmt.Errorf("Invalid address %s (%s)", addr, err)
		}
		l, err := net.ListenTCP("tcp", a)
		if err != nil {
			return err
		}
		f, err := l.File()
		if err != nil {
			return fmt.Errorf("Failed to retreive fd for: %s (%s)", addr, err)
		}
		if err := l.Close(); err != nil {
			return fmt.Errorf("Failed to close listener for: %s (%s)", addr, err)
		}
		mp.slaveExtraFiles[i] = f
	}
	return nil
}

//fetchLoop is run in a goroutine
func (mp *master) fetchLoop() {
	min := mp.Config.MinFetchInterval
	time.Sleep(min)
	for {
		t0 := time.Now()
		mp.fetch()
		//duration fetch of fetch
		diff := time.Now().Sub(t0)
		if diff < min {
			delay := min - diff
			//ensures at least MinFetchInterval delay.
			//should be throttled by the fetcher!
			time.Sleep(delay)
		}
	}
}

func (mp *master) fetch() {
	if mp.printCheckUpdate {
		mp.logger.Debugf("checking for updates...")
	}
	reader, err := mp.Fetcher.Fetch()
	if err != nil {
		mp.logger.Debugf("failed to get latest version: %s", err)
		return
	}
	if reader == nil {
		if mp.printCheckUpdate {
			mp.logger.Debugf("no updates")
		}
		mp.printCheckUpdate = false
		return //fetcher has explicitly said there are no updates
	}
	mp.printCheckUpdate = true
	mp.logger.Debugf("streaming update...")
	//optional closer
	if closer, ok := reader.(io.Closer); ok {
		defer closer.Close()
	}
	tmpBin, err := os.OpenFile(tmpBinPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		mp.logger.Errorf("failed to open temp binary: %s", err)
		return
	}
	defer func() {
		tmpBin.Close()
		os.Remove(tmpBinPath)
	}()
	//tee off to sha1
	hash := sha1.New()
	reader = io.TeeReader(reader, hash)
	//write to a temp file
	_, err = io.Copy(tmpBin, reader)
	if err != nil {
		mp.logger.Errorf("failed to write temp binary: %s", err)
		return
	}
	//compare hash
	newHash := hash.Sum(nil)
	if bytes.Equal(mp.binHash, newHash) {
		mp.logger.Debugf("hash match - skip")
		return
	}
	//copy permissions
	if err := chmod(tmpBin, mp.binPerms); err != nil {
		mp.logger.Errorf("failed to make temp binary executable: %s", err)
		return
	}
	if err := chown(tmpBin, uid, gid); err != nil {
		mp.logger.Errorf("failed to change owner of binary: %s", err)
		return
	}
	if _, err := tmpBin.Stat(); err != nil {
		mp.logger.Errorf("failed to stat temp binary: %s", err)
		return
	}
	tmpBin.Close()
	if _, err := os.Stat(tmpBinPath); err != nil {
		mp.logger.Errorf("failed to stat temp binary by path: %s", err)
		return
	}
	if mp.Config.PreUpgrade != nil {
		if err := mp.Config.PreUpgrade(tmpBinPath); err != nil {
			mp.logger.Errorf("user cancelled upgrade: %s", err)
			return
		}
	}
	//overseer sanity check, dont replace our good binary with a non-executable file
	tokenIn := token()
	cmd := exec.Command(tmpBinPath)
	cmd.Env = append(os.Environ(), []string{envBinCheck + "=" + tokenIn}...)
	cmd.Args = os.Args
	returned := false
	go func() {
		time.Sleep(5 * time.Second)
		if !returned {
			mp.logger.Warnf("sanity check against fetched executable timed-out, check overseer is running")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
	}()
	tokenOut, err := cmd.CombinedOutput()
	returned = true
	if err != nil {
		mp.logger.Errorf("failed to run temp binary: %s (%s) output \"%s\"", err, tmpBinPath, tokenOut)
		return
	}
	if tokenIn != string(tokenOut) {
		mp.logger.Warnf("sanity check failed")
		return
	}
	//overwrite!
	if err := move(mp.binPath, tmpBinPath); err != nil {
		mp.logger.Errorf("failed to overwrite binary: %s", err)
		return
	}
	mp.logger.Debugf("upgraded binary (%x -> %x)", mp.binHash[:12], newHash[:12])
	mp.binHash = newHash
	//binary successfully replaced
	if !mp.Config.NoRestartAfterFetch {
		mp.restart.Run()
	}
	//and keep fetching...
	return
}

func token() string {
	buff := make([]byte, 8)
	rand.Read(buff)
	return hex.EncodeToString(buff)
}
