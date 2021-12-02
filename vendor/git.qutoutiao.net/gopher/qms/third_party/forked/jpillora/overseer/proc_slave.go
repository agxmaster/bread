package overseer

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	//DisabledState is a placeholder state for when
	//overseer is disabled and the program function
	//is run manually.
	DisabledState = State{Enabled: false}
)

// State contains the current run-time state of overseer
type State struct {
	//whether overseer is running enabled. When enabled,
	//this program will be running in a child process and
	//overseer will perform rolling upgrades.
	Enabled bool
	//ID is a SHA-1 hash of the current running binary
	ID string
	//StartedAt records the start time of the program
	StartedAt time.Time
	//Listener is the first net.Listener in Listeners
	Listener net.Listener
	//Listeners are the set of acquired sockets by the master
	//process. These are all passed into this program in the
	//same order they are specified in Config.Addresses.
	Listeners []net.Listener
	//Program's first listening address
	Address string
	//Program's listening addresses
	Addresses []string
	//GracefulShutdown will be filled when its time to perform
	//a graceful shutdown.
	gracefulC chan struct{}
	// 程序准备就绪
	programReadyC chan struct{}
	//Path of the binary currently being executed
	BinPath string
}

// 当前slave已经ready
func (s *State) ProgramReady() {
	if s.programReadyC != nil {
		s.programReadyC <- struct{}{}
	}
}

func (s *State) Graceful() <-chan struct{} {
	if s.gracefulC == nil {
		s.gracefulC = make(chan struct{}, 1)
	}
	return s.gracefulC
}

//a overseer slave process
type slave struct {
	*Config
	logger
	id         string
	listeners  []*overseerListener
	masterPid  int
	masterProc *os.Process
	state      State
}

func newSlave(cfg *Config) *slave {
	sp := &slave{
		Config:    cfg,
		id:        os.Getenv(envSlaveID),
		masterPid: os.Getppid(),
		state: State{
			Enabled:       true,
			ID:            os.Getenv(envBinID),
			Address:       cfg.Address,
			Addresses:     cfg.Addresses,
			gracefulC:     make(chan struct{}, 1),
			programReadyC: make(chan struct{}, 1),
			BinPath:       os.Getenv(envBinPath),
		},
	}
	sp.logger = newWrapLogger(fmt.Sprintf("[overseer slave#%s] ", sp.id), cfg.Logger)
	return sp
}

func (sp *slave) run() error {
	sp.logger.Debugf("run")
	sp.state.StartedAt = time.Now()
	if err := sp.watchParent(); err != nil {
		return err
	}
	if err := sp.initFileDescriptors(); err != nil {
		return err
	}
	sp.watchSignal()
	go sp.watchProgramReady()
	//run program with state
	sp.logger.Debugf("start program")
	sp.Config.Program(sp.state)
	return nil
}

func (sp *slave) watchParent() error {
	proc, err := os.FindProcess(sp.masterPid)
	if err != nil {
		return fmt.Errorf("master process: %s", err)
	}
	sp.masterProc = proc
	go func() {
		//send signal 0 to master process forever
		for {
			//should not error as long as the process is alive
			if err := sp.masterProc.Signal(syscall.Signal(0)); err != nil {
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
		}
	}()
	return nil
}

func (sp *slave) initFileDescriptors() error {
	//inspect file descriptors
	numFDs, err := strconv.Atoi(os.Getenv(envNumFDs))
	if err != nil {
		return fmt.Errorf("invalid %s integer", envNumFDs)
	}
	sp.listeners = make([]*overseerListener, numFDs)
	sp.state.Listeners = make([]net.Listener, numFDs)
	for i := 0; i < numFDs; i++ {
		f := os.NewFile(uintptr(3+i), "")
		l, err := net.FileListener(f)
		if err != nil {
			return fmt.Errorf("failed to inherit file descriptor: %d", i)
		}
		u := newOverseerListener(l)
		sp.listeners[i] = u
		sp.state.Listeners[i] = u
	}
	if len(sp.state.Listeners) > 0 {
		sp.state.Listener = sp.state.Listeners[0]
	}
	return nil
}

func (sp *slave) watchSignal() {
	signals := make(chan os.Signal)
	signal.Notify(signals, sp.Config.RestartSignal)
	go func() {
		<-signals
		signal.Stop(signals)
		sp.logger.Tracef("graceful shutdown requested")
		sp.graceful()
	}()
}

func (sp *slave) watchProgramReady() {
	<-sp.state.programReadyC
	if err := sp.masterProc.Signal(SIGUSR1); err != nil {
		sp.logger.Errorf("send signal[USR1] to master failed: %v", err)
		os.Exit(1)
	}
}

// restart step-2 close old slave
func (sp *slave) graceful() {
	close(sp.state.gracefulC)
	time.Sleep(50 * time.Millisecond) // 务必让业务端先处理graceful channel
	// 释放socket连接
	if len(sp.listeners) > 0 {
		//perform graceful shutdown
		for _, l := range sp.listeners {
			l.release(sp.Config.TerminateTimeout)
		}
	}
	//start death-timer
	go func() {
		time.Sleep(sp.Config.TerminateTimeout)
		sp.logger.Warnf("timeout. forceful shutdown")
		os.Exit(1)
	}()
}
