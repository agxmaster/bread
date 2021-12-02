package cc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/magiconair/properties"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"git.qutoutiao.net/gopher/cc-client-go/log"
	"git.qutoutiao.net/gopher/cc-client-go/proto-gen/admin_sdk"
)

// NewNullConfigCenter 参数失败初始化的cc对象
func NewNullConfigCenter(projectName string, projectToken string, env Env) *ConfigCenter {
	return &ConfigCenter{
		ProjectName:               projectName,
		env:                       env,
		data:                      make(map[string]string),
		dataLock:                  new(sync.RWMutex),
		appLogger:                 log.StdLogger,
		diagnosticLoggers:         [2]log.Logger{log.NullLogger, log.NullLogger},
		latestConfigVariableTagId: 0,
		latestPublishTimestamp:    0,
	}
}

type ConfigCenter struct {
	ProjectName string
	env         Env
	// 变量的数据
	data map[string]string
	// 变量获取的锁
	dataLock *sync.RWMutex

	appLogger         log.Logger
	diagnosticLoggers [2]log.Logger

	// 同步最新的tagID
	latestConfigVariableTagId int64
	// 同步最新的发布时间戳
	latestPublishTimestamp int64
	// 备份目录, 文件名称({projectName}_{env}.json)
	backupDir      string
	backupFilePath string
	// 根据是否存在checksum文件来判断是否对数据checksum
	checksumFilePath string
	checksumNeeded   bool

	streamCallOpts []grpc.CallOption
	unaryCallOpts  []grpc.CallOption

	sdk            *admin_sdk.SDKClient
	retryableCodes []codes.Code

	context context.Context
	cancel  context.CancelFunc
	closed  int32

	conn        *grpc.ClientConn
	connLock    *sync.RWMutex
	client      admin_sdk.AdminSDKClient
	stream      admin_sdk.AdminSDK_PushVariablesClient
	onChange    func(*ConfigCenter) error
	bibUpdater  Updater
	tickUpdater Updater
	// 双向流重试次数
	attempt        uint64
	backoffOptions *backoffOptions

	// debug
	debug bool
	// qa环境地址可以自定义，prd环境不可自定义
	QAServerURL string
}

type backoffOptions struct {
	backoffFunc     backoffFunc
	MaxInterval     time.Duration
	InitialInterval time.Duration
	JitterFraction  float64
}

func (c ConfigCenter) DebugInfo() string {
	latestPublishTimestamp := atomic.LoadInt64(&c.latestPublishTimestamp)
	latestConfigVariableTagId := atomic.LoadInt64(&c.latestConfigVariableTagId)
	return fmt.Sprintf("projectName: %v, env: %v, backup file path: %v, lastest publish timestamp: %v, latest config variable tagID: %v",
		c.ProjectName, c.env, c.backupFilePath, latestPublishTimestamp, latestConfigVariableTagId)
}

func (c *ConfigCenter) Close() error {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		return nil
	}
	for _, logger := range c.diagnosticLoggers {
		if _, ok := logger.(io.Closer); ok {
			logger.Close()
		}
	}
	if _, ok := c.appLogger.(io.Closer); ok {
		c.appLogger.Close()
	}
	if c.cancel != nil {
		c.cancel()
		if c.debug {
			stdlog.Printf("%v cancel", CCClientVersion)
		}
	}
	if c.debug {
		stdlog.Printf("%v closed.", CCClientVersion)
	}
	return nil
}

func (c *ConfigCenter) GetString(key, defaultVal string) string {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	if val, ok := c.data[key]; ok {
		return val
	}
	c.appLogger.Error(key)
	return defaultVal
}

func (c *ConfigCenter) GetInt(key string, defaultVal int) int {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	if val, ok := c.data[key]; ok {
		intVal, err := strconv.Atoi(val)
		if err == nil {
			return intVal
		}
		return defaultVal
	} else {
		c.appLogger.Error(key)
		return defaultVal
	}
}

func (c *ConfigCenter) GetFloat(key string, defaultVal float64) float64 {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	if val, ok := c.data[key]; ok {
		floatVal, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return floatVal
		}
		return defaultVal
	} else {
		c.appLogger.Error(key)
		return defaultVal
	}
}

func (c *ConfigCenter) GetBool(key string, defaultVal bool) bool {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	if val, ok := c.data[key]; ok {
		boolVal, err := strconv.ParseBool(val)
		if err == nil {
			return boolVal
		}
		return defaultVal
	} else {
		c.appLogger.Error(key)
		return defaultVal
	}
}

func (c *ConfigCenter) GetAll() map[string]string {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	res := make(map[string]string)
	for k, v := range c.data {
		res[k] = v
	}
	return res
}

func (c *ConfigCenter) IsKeyExists(key string) bool {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	_, ok := c.data[key]
	return ok
}

func (c *ConfigCenter) needUpdate(info *admin_sdk.VariableInfo) bool {
	return atomic.LoadInt64(&c.latestPublishTimestamp) < info.PublishTimestamp
}

func (c *ConfigCenter) updateVariables(info *admin_sdk.VariableInfo) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	if c.debug {
		stdlog.Printf("before: %v, variables: %v", c.data, info.GetVariables())
	}
	atomic.StoreInt64(&c.latestConfigVariableTagId, info.ConfigVariableTagId)
	atomic.StoreInt64(&c.latestPublishTimestamp, info.PublishTimestamp)
	if !info.GetModified() {
		return
	}
	for k := range c.data {
		delete(c.data, k)
	}
	for _, v := range info.GetVariables() {
		c.data[v.ConfigVariableKey] = v.ConfigVariableValue
	}
	if c.debug {
		stdlog.Printf("after: %v", c.data)
	}
}

func (c *ConfigCenter) makeBackupDir() error {
	return os.MkdirAll(c.backupDir, 0700)
}

func (c *ConfigCenter) createChecksumPath() error {
	_, err := os.OpenFile(c.checksumFilePath, os.O_RDWR|os.O_CREATE, 0666)
	return err
}

func (c *ConfigCenter) backupVariables(info *admin_sdk.VariableInfo) error {
	if err := c.makeBackupDir(); err != nil {
		return err
	}
	fd, err := os.OpenFile(c.backupFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()
	encoder := json.NewEncoder(fd)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(info)
	if err != nil {
		for _, logger := range c.diagnosticLoggers {
			logger.Errorf("backupVariables failed: %v", err)
		}
	}
	return c.createChecksumPath()
}

func (c *ConfigCenter) restoreVariables(info *admin_sdk.VariableInfo) error {
	fd, err := os.Open(c.backupFilePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	decoder := json.NewDecoder(fd)
	err = decoder.Decode(info)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *ConfigCenter) waitRetryBackoff() error {
	var waitTime time.Duration = 0
	attempt := atomic.LoadUint64(&c.attempt)
	if attempt > 0 {
		waitTime = c.backoffOptions.backoffFunc(attempt)
	}
	if waitTime > 0 {
		if c.debug {
			stdlog.Printf("grpc retry backoff for %v", waitTime)
		}
		after := time.After(waitTime)
		select {
		case <-after:
		case <-c.context.Done():
			return c.context.Err()
		}
	}
	return nil
}

func (c *ConfigCenter) pullVariables(ctx context.Context) (*admin_sdk.VariableInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	in := &admin_sdk.GetVariablesReq{
		Client:              c.sdk,
		ConfigVariableTagId: atomic.LoadInt64(&c.latestConfigVariableTagId),
	}
	client := c.getClient()
	res, err := client.GetVariables(ctx, in, c.unaryCallOpts...)
	if err != nil {
		for _, logger := range c.diagnosticLoggers {
			logger.Errorf("GetVariables Error: %v", err)
		}
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.New(res.Desc)
	}
	return res.GetVariableInfo(), err
}

func (c *ConfigCenter) closeConn() error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ConfigCenter) getConn() *grpc.ClientConn {
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	return c.conn
}

func (c *ConfigCenter) resetConn() error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	conn, err := NewGRPCConn(ctx, c.env, c.QAServerURL)
	if err != nil {
		return err
	}
	c.conn = conn
	client := admin_sdk.NewAdminSDKClient(conn)
	c.client = client
	stream, err := client.PushVariables(context.Background(), c.streamCallOpts...)
	if err != nil {
		return err
	}
	c.stream = stream
	return nil
}

func (c *ConfigCenter) getClient() admin_sdk.AdminSDKClient {
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	return c.client
}

func (c *ConfigCenter) getStream() admin_sdk.AdminSDK_PushVariablesClient {
	c.connLock.RLock()
	defer c.connLock.RUnlock()
	return c.stream
}

func projectNameNormalize(projectName string) string {
	return strings.Replace(projectName, "-", "_", -1)
}

func NewDevConfigCenter(projectName string, configPath string, options ...Option) (cc *ConfigCenter, err error) {
	projectName = projectNameNormalize(projectName)
	// 1. 配置加载
	opts := Options{
		diagnosticLogger:    [2]log.Logger{},
		restoreWhenInitFail: true,
	}
	for _, option := range options {
		option(&opts)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cc = &ConfigCenter{
		ProjectName:       projectName,
		env:               DEV,
		data:              make(map[string]string),
		dataLock:          new(sync.RWMutex),
		appLogger:         nil,
		diagnosticLoggers: [2]log.Logger{},
		connLock:          new(sync.RWMutex),
		context:           ctx,
		cancel:            cancel,
		backoffOptions: &backoffOptions{
			MaxInterval:     DefaultMaxInterval,
			InitialInterval: DefaultInitialInterval,
			JitterFraction:  DefaultJitterFraction,
		},
	}
	cc.debug = opts.debug
	if configPath == "" {
		return cc, errors.New("配置中心sdk开发环境是通过读取的本地配置文件加载，但是配置文件不存在")
	}
	bs, err := ioutil.ReadFile(configPath)
	if err != nil {
		return cc, err
	}
	p, err := properties.LoadString(string(bs))
	if err != nil {
		return cc, err
	}
	m := p.Map()
	cc.dataLock.Lock()
	for k := range cc.data {
		delete(cc.data, k)
	}
	for k, v := range m {
		cc.data[k] = v
	}
	defer cc.dataLock.Unlock()
	return cc, nil
}

func NewConfigCenter(projectName string, projectToken string, env Env, options ...Option) (cc *ConfigCenter, err error) {
	projectName = projectNameNormalize(projectName)
	// 1. 配置加载
	opts := Options{
		diagnosticLogger:    [2]log.Logger{},
		restoreWhenInitFail: true,
	}
	for _, option := range options {
		option(&opts)
	}
	if env == DEV && opts.DevConfigFilePath != "" {
		return NewDevConfigCenter(projectName, opts.DevConfigFilePath, options...)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cc = &ConfigCenter{
		ProjectName:       projectName,
		env:               env,
		data:              make(map[string]string),
		dataLock:          new(sync.RWMutex),
		appLogger:         nil,
		diagnosticLoggers: [2]log.Logger{},
		connLock:          new(sync.RWMutex),
		context:           ctx,
		cancel:            cancel,
		backoffOptions: &backoffOptions{
			MaxInterval:     DefaultMaxInterval,
			InitialInterval: DefaultInitialInterval,
			JitterFraction:  DefaultJitterFraction,
		},
	}
	cc.debug = opts.debug

	cc.QAServerURL = opts.QAServerURL
	if cc.QAServerURL == "" {
		cc.QAServerURL = QAServerAddr
	}
	ipString := opts.clientIP
	if ipString == "" {
		if ip, ipErr := HostIP(); ipErr != nil {
			ipString = ""
		} else {
			ipString = ip.String()
		}
	}
	var hostName string
	var hostNameErr error
	if hostName, hostNameErr = os.Hostname(); hostNameErr != nil {
		hostName = ""
	}
	var processName string
	var processNameErr error
	if processName, processNameErr = CurrentProcessName(); processNameErr != nil {
		projectName = ""
	}
	sdkInfo := &admin_sdk.SDKClient{
		Ip:           ipString,
		UserAgent:    CCClientVersion,
		Env:          string(env),
		ProjectName:  projectName,
		ProcessName:  processName,
		ProjectToken: projectToken,
		HostName:     hostName,
	}
	cc.sdk = sdkInfo

	cc.backupDir = opts.backupDir
	if len(cc.backupDir) == 0 {
		cc.backupDir = fmt.Sprintf(QADefaultBackupDir, cc.ProjectName)
	}
	if mkErr := cc.makeBackupDir(); mkErr != nil {
		err = errors.WithStack(mkErr)
	}
	gRPCctx, gRPCcancel := context.WithTimeout(context.Background(), GRPCTimeout)
	defer gRPCcancel()
	conn, connErr := NewGRPCConn(gRPCctx, env, cc.QAServerURL)
	if connErr != nil {
		if err == nil {
			err = errors.WithStack(connErr)
		}
	}
	cc.conn = conn
	cc.client = admin_sdk.NewAdminSDKClient(conn)
	cc.diagnosticLoggers = opts.diagnosticLogger
	if cc.diagnosticLoggers[0] == nil {
		cc.diagnosticLoggers[0] = log.NullLogger
	}
	cc.diagnosticLoggers[1] = NewLogger(cc, cc.sdk)
	appLogger := NewAppLogger(cc, cc.sdk)
	appLogger.SetCenter(cc)
	cc.appLogger = appLogger

	cc.retryableCodes = opts.retryableCodes
	cc.backoffOptions.backoffFunc = backoffExponentialWithJitter(cc.backoffOptions.JitterFraction, cc.backoffOptions.InitialInterval, cc.backoffOptions.MaxInterval)
	if len(cc.streamCallOpts) == 0 {
		cc.streamCallOpts = append(cc.streamCallOpts, grpc.WaitForReady(true))
	}
	cc.backoffOptions.JitterFraction = opts.jitterFraction
	if cc.backoffOptions.JitterFraction == 0 {
		cc.backoffOptions.JitterFraction = DefaultJitterFraction
	}
	cc.backoffOptions.MaxInterval = opts.maxInterval
	if cc.backoffOptions.MaxInterval == 0 {
		cc.backoffOptions.MaxInterval = DefaultMaxInterval
	}
	cc.backoffOptions.InitialInterval = opts.initialInterval
	if cc.backoffOptions.InitialInterval == 0 {
		cc.backoffOptions.InitialInterval = DefaultInitialInterval
	}

	cc.onChange = opts.onChange
	if opts.onChange != nil {
		cc.onChange = DoWithTimeoutClosure(opts.onChange, CallbackTimeout)
	}
	cc.backupFilePath = filepath.Join(cc.backupDir, cc.ProjectName+"_"+string(cc.env)+".json")
	cc.checksumFilePath = filepath.Join(cc.backupDir, cc.ProjectName+"_"+string(cc.env)+".checksum")
	cc.checksumNeeded = func() bool {
		if _, err := os.Stat(cc.checksumFilePath); os.IsNotExist(err) {
			return false
		}
		return true
	}()
	info := &admin_sdk.VariableInfo{}
	restoreErr := cc.restoreVariables(info)
	var checkSumError error
	if restoreErr == nil {
		if cc.checksumNeeded {
			calcCheckSum := checkSum(info.GetVariables())
			if calcCheckSum != info.GetCheckSum() {
				checkSumError = ErrChecksum{info.CheckSum, calcCheckSum}
				if cc.debug {
					stdlog.Printf("restore backup file failed: %v", checkSumError)
				}
			} else {
				cc.updateVariables(info)
			}
		} else {
			cc.updateVariables(info)
		}
	}
	info, pullErr := cc.pullVariables(ctx)
	if pullErr != nil {
		err = pullErr
		/*
			if opts.restoreWhenInitFail {
				if restoreErr != nil {
					err = restoreErr
				} else if checkSumError != nil {
					err = checkSumError
				}
			}
		*/
	} else if info != nil && len(info.Variables) > 0 {
		calcCheckSum := checkSum(info.Variables)
		if info.CheckSum != calcCheckSum {
			if err == nil {
				err = errors.WithStack(ErrChecksum{info.CheckSum, calcCheckSum})
			}
		} else {
			// 3. 更新变量数据
			if cc.needUpdate(info) {
				cc.updateVariables(info)
				if cc.onChange != nil {
					if onChangeErr := cc.onChange(cc); onChangeErr != nil {
						if err == nil {
							err = onChangeErr
						}
					}
				}
				// 4. 备份数据
				if bakErr := cc.backupVariables(info); bakErr != nil {
					if err == nil {
						err = ErrBackup
					}
				}
			}
		}
	}
	cc.bibUpdater = NewDuplexUpdate(newPushVariableSender(cc))
	cc.tickUpdater = &TickerUpdate{}
	go cc.initStream()
	return
}

func (c *ConfigCenter) initStream() {
	stream, err := c.client.PushVariables(context.Background(), c.streamCallOpts...)
	if c.debug {
		if err != nil {
			stdlog.Printf("%v establish connection failed", CCClientVersion)
		} else {
			stdlog.Printf("%v establish connection successfully", CCClientVersion)
		}
	}
	c.connLock.Lock()
	c.stream = stream
	c.connLock.Unlock()
	go c.bibUpdater.update(context.Background(), c)
	c.tickUpdater.update(context.Background(), c)
}

func checkSum(variables []*admin_sdk.Variable) string {
	// 1. sort by key
	sort.Slice(variables, func(i, j int) bool {
		return variables[i].ConfigVariableId < variables[j].ConfigVariableId
	})
	// 2. arrange by k1v1t1k2v1t2
	var raw string
	for _, v := range variables {
		raw += v.ConfigVariableKey + v.ConfigVariableValue + v.ConfigVariableValueType
	}
	// 3. md5 checksum
	return MD5CheckSum(raw)
}
