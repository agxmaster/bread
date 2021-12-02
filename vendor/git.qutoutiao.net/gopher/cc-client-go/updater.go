package cc

import (
	"context"
	"io"
	"log"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"git.qutoutiao.net/gopher/cc-client-go/proto-gen/admin_sdk"
)

type Updater interface {
	update(ctx context.Context, center *ConfigCenter)
}

type DuplexUpdate struct {
	Sender
}

func NewDuplexUpdate(sender Sender) *DuplexUpdate {
	return &DuplexUpdate{sender}
}

type Sender interface {
	SendConnect() error
	SendReceiveOk(string) error
	SendReloadOk(string) error
	SendCallBackOk(string) error
	SendCallBackSkip(string) error
	SendPersistenceOk(string) error
	SendReceiveFail(string) error
	SendReloadFail(string, error) error
	SendCallBackFail(string, error) error
	SendPersistenceFail(string, error) error
	SendHeartbeat() error
	SendDisConnect() error
	Recv() (*admin_sdk.PushVariablesResp, error)
}

type pushVariablesSender struct {
	center *ConfigCenter
}

func newPushVariableSender(center *ConfigCenter) *pushVariablesSender {
	return &pushVariablesSender{center: center}
}

func SendWithTimeout(f func(req *admin_sdk.PushVariablesReq) error, req *admin_sdk.PushVariablesReq, timeout time.Duration) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- f(req)
		close(errChan)
	}()
	t := time.NewTimer(timeout)
	select {
	case <-t.C:
		return status.Errorf(codes.DeadlineExceeded, "请求超时")
	case err := <-errChan:
		if !t.Stop() {
			<-t.C
		}
		return err
	}
}

func (p pushVariablesSender) SendConnect() error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_Connect,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
		PublishTimestamp:    atomic.LoadInt64(&p.center.latestPublishTimestamp),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendReceiveOk(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_ReceiveOk,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendReloadOk(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_ReloadOk,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendCallBackOk(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_CallBackOk,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	if p.center.debug {
		log.Printf("requestID: %v callback ok.", requestID)
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendCallBackSkip(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_CallBackSkip,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	if p.center.debug {
		log.Printf("requestID: %v callback skip.", requestID)
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendPersistenceOk(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_PersistenceOk,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendReceiveFail(requestID string) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_ReceiveFail,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendReloadFail(requestID string, err error) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_ReloadFail,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
		Desc:                err.Error(),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendCallBackFail(requestID string, err error) error {
	desc := ""
	if err != nil {
		desc = err.Error()
	}
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_CallBackFail,
		RequestId:           requestID,
		Desc:                desc,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendPersistenceFail(requestID string, err error) error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_PersistenceFail,
		RequestId:           requestID,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
		Desc:                err.Error(),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendHeartbeat() error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_Heartbeat,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) SendDisConnect() error {
	var req = &admin_sdk.PushVariablesReq{
		Client:              p.center.sdk,
		Instruction:         admin_sdk.SDKClientInstruction_DisConnect,
		ConfigVariableTagId: atomic.LoadInt64(&p.center.latestConfigVariableTagId),
	}
	return SendWithTimeout(p.center.getStream().Send, req, gRPCSendTimeout)
}

func (p pushVariablesSender) Recv() (*admin_sdk.PushVariablesResp, error) {
	return p.center.getStream().Recv()
}

func (d DuplexUpdate) instructionLogError(center *ConfigCenter, err error, instruction admin_sdk.SDKClientInstruction, requestID string) {
	for _, logger := range center.diagnosticLoggers {
		logger.Errorf("info: [%v], requestID: [%v], instruction: [%v], err: [%v], connection: [%p], timestamp: [%v]",
			center.DebugInfo(), requestID, instruction, err, center.getConn(), time.Now().Format("2006-01-02 15:04:05"))
	}
}

func (d DuplexUpdate) instructionLogInfo(center *ConfigCenter, instruction admin_sdk.SDKClientInstruction, requestID string) {
	switch instruction {
	case admin_sdk.SDKClientInstruction_Connect:
		for _, logger := range center.diagnosticLoggers {
			logger.Infof("info: [%v], connect with cc server success, connection: [%p], timestamp: [%v]",
				center.DebugInfo(), center.getConn(), time.Now().Format("2006-01-02 15:04:05"))
		}
	default:
		for _, logger := range center.diagnosticLoggers {
			logger.Infof("info: [%v], requestID: [%v], instruction: [%v], connection: [%p], timestamp: [%v]",
				center.DebugInfo(), requestID, instruction, center.getConn(), time.Now().Format("2006-01-02 15:04:05"))
		}
	}
}

func (d DuplexUpdate) logPanic(center *ConfigCenter, r interface{}) {
	for _, logger := range center.diagnosticLoggers {
		logger.Errorf("info: [%v], connection: [%p], timestamp: [%v], panic: %v",
			center.DebugInfo(), center.getConn(), time.Now().Format("2006-01-02 15:04:05"), r)
	}

}

func (d DuplexUpdate) heartbeat(center *ConfigCenter, stop <-chan struct{}) {
	go func() {
		ticker := time.NewTicker(HeartbeatInterval)
		defer ticker.Stop()
		var err error
		for {
			select {
			case <-ticker.C:
				err = d.SendHeartbeat()
				if err == nil {
					continue
				}
				if center.debug {
					log.Printf("send heartbeat failed, connection: %p", center.getConn())
				}
				if err == io.EOF || isRetryable(err, center.retryableCodes) {
					center.closeConn()
					atomic.AddUint64(&(center.attempt), 1)
				}
				return
			case <-stop:
				return
			}
		}
	}()
}

func (d DuplexUpdate) ReceiveStage(center *ConfigCenter, resp *admin_sdk.PushVariablesResp, heartbeatStop chan struct{}) bool {
	d.instructionLogInfo(center, admin_sdk.SDKClientInstruction_ReceiveOk, resp.RequestId)
	err := d.SendReceiveOk(resp.RequestId)
	if center.debug {
		log.Printf("requestID: %v, send receive ok, connection: %p", resp.RequestId, center.getConn())
	}
	if err != nil {
		if err == io.EOF || isRetryable(err, center.retryableCodes) {
			close(heartbeatStop)
			atomic.AddUint64(&(center.attempt), 1)
			return true
		}
		d.instructionLogError(center, err, admin_sdk.SDKClientInstruction_ReceiveOk, resp.RequestId)
	}
	return false
}

func (d DuplexUpdate) ReloadStage(center *ConfigCenter, resp *admin_sdk.PushVariablesResp, heartbeatStop chan struct{}) (bk bool, checksum bool, needUpdate bool) {
	info := resp.GetVariableInfo()
	calCheckSum := checkSum(info.GetVariables())
	if calCheckSum != info.GetCheckSum() {
		err := ErrChecksum{calCheckSum: calCheckSum, checksum: info.GetCheckSum()}
		d.SendReloadFail(resp.GetRequestId(), err)
		d.instructionLogError(center, err, admin_sdk.SDKClientInstruction_ReloadFail, resp.RequestId)
		return
	}
	checksum = true
	needUpdate = center.needUpdate(info)
	if needUpdate {
		center.updateVariables(info)
	}
	d.instructionLogInfo(center, admin_sdk.SDKClientInstruction_ReloadOk, resp.RequestId)
	err := d.SendReloadOk(resp.RequestId)
	if err != nil {
		if err == io.EOF || isRetryable(err, center.retryableCodes) {
			close(heartbeatStop)
			atomic.AddUint64(&(center.attempt), 1)
			bk = true
			return
		}
		d.instructionLogError(center, err, admin_sdk.SDKClientInstruction_ReloadOk, resp.RequestId)
	}
	return
}

// TODO 函数签名
func (d DuplexUpdate) CallbackStage(center *ConfigCenter, resp *admin_sdk.PushVariablesResp, heartbeatStop chan struct{}, needUpdate bool) bool {
	callbackInstruction := admin_sdk.SDKClientInstruction_CallBackOk
	var err error
	if center.onChange == nil || !needUpdate {
		err = d.SendCallBackSkip(resp.RequestId)
		callbackInstruction = admin_sdk.SDKClientInstruction_CallBackSkip
	} else {
		if callbackErr := center.onChange(center); callbackErr != nil {
			callbackInstruction = admin_sdk.SDKClientInstruction_CallBackOk
			err = d.SendCallBackFail(resp.RequestId, callbackErr)
		} else {
			err = d.SendCallBackOk(resp.RequestId)
		}
	}
	d.instructionLogInfo(center, callbackInstruction, resp.RequestId)
	if err != nil {
		if err == io.EOF || isRetryable(err, center.retryableCodes) {
			close(heartbeatStop)
			atomic.AddUint64(&(center.attempt), 1)
			return true
		}
		d.instructionLogError(center, err, callbackInstruction, resp.RequestId)
	}
	return false
}

func (d DuplexUpdate) PersistenceStage(center *ConfigCenter, resp *admin_sdk.PushVariablesResp, heartbeatStop chan struct{}, needUpdate bool) bool {
	var err error
	persistenceInstruction := admin_sdk.SDKClientInstruction_PersistenceOk
	if !needUpdate {
		err = d.SendPersistenceOk(resp.RequestId)
	} else {
		if err = center.backupVariables(resp.VariableInfo); err != nil {
			persistenceInstruction = admin_sdk.SDKClientInstruction_PersistenceFail
			err = d.SendPersistenceFail(resp.RequestId, err)
		} else {
			err = d.SendPersistenceOk(resp.RequestId)
		}
	}
	d.instructionLogInfo(center, persistenceInstruction, resp.RequestId)
	if err != nil {
		if err == io.EOF || isRetryable(err, center.retryableCodes) {
			close(heartbeatStop)
			atomic.AddUint64(&(center.attempt), 1)
			return true
		}
		d.instructionLogError(center, err, persistenceInstruction, resp.RequestId)
	}
	return false
}

// 推送变量到sdk
func (d DuplexUpdate) update(ctx context.Context, center *ConfigCenter) {
	defer func() {
		if err := recover(); err != nil {
			d.logPanic(center, err)
		}
	}()
	var err error
	var firstReq = true
	for {
		if !firstReq {
			if err = center.waitRetryBackoff(); err != nil {
				return
			}
		} else {
			firstReq = false
		}
		err = d.SendConnect()
		if err != nil {
			if err == io.EOF || isRetryable(err, center.retryableCodes) {
				center.resetConn()
				atomic.AddUint64(&(center.attempt), 1)
				continue
			}
			d.instructionLogError(center, err, admin_sdk.SDKClientInstruction_Connect, "")
			continue
		}
		d.instructionLogInfo(center, admin_sdk.SDKClientInstruction_Connect, "")
		if center.debug {
			log.Printf("%v send connect instruction success, connection: %p", CCClientVersion, center.getConn())
		}
		var resp = &admin_sdk.PushVariablesResp{}
		heartbeatStop := make(chan struct{})
		d.heartbeat(center, heartbeatStop)
	RECONNECT:
		for {
			select {
			case <-ctx.Done():
				d.SendDisConnect()
				return
			default:
				if center.debug {
					log.Printf("sdk: %v, wait for receive, connection: %p", center.sdk, center.getConn())
				}
				resp, err = d.Recv()
				if center.debug {
					log.Printf("sdk: %v, receive: %v, err: %v, connection: %p", center.sdk, resp, err, center.getConn())
				}
				if err != nil {
					if err == io.EOF || isRetryable(err, center.retryableCodes) {
						close(heartbeatStop)
						atomic.AddUint64(&(center.attempt), 1)
						break RECONNECT
					}
					if resp != nil {
						d.SendReceiveFail(resp.GetRequestId())
					}
					d.instructionLogError(center, err, admin_sdk.SDKClientInstruction_UnKnow, resp.GetRequestId())
					continue
				}
				if resp.GetCode() == 0 {
					atomic.StoreUint64(&(center.attempt), 0)
				}
				switch resp.GetInstruction() {
				case admin_sdk.AdminInstruction_Publish:
					// ---- 接收阶段
					if d.ReceiveStage(center, resp, heartbeatStop) {
						break RECONNECT
					}
					// ---- 重载阶段
					bk, cs, nu := d.ReloadStage(center, resp, heartbeatStop)
					if bk {
						break RECONNECT
					}
					if !cs {
						break
					}
					// ---- 回调阶段
					if d.CallbackStage(center, resp, heartbeatStop, nu) {
						break RECONNECT
					}
					// ---- 备份阶段
					if d.PersistenceStage(center, resp, heartbeatStop, nu) {
						break RECONNECT
					}
				case admin_sdk.AdminInstruction_HeartbeatAck:
					if center.debug {
						log.Printf("rquestID: %v, recv: %v, connection: %p", resp.RequestId, admin_sdk.AdminInstruction_HeartbeatAck, center.getConn())
					}
				case admin_sdk.AdminInstruction_ServerDisConnect:
					if center.debug {
						log.Printf("rquestID: %v, recv: %v, connection: %p", resp.RequestId, admin_sdk.AdminInstruction_ServerDisConnect, center.getConn())
					}
					center.closeConn()
					// 如果是业务逻辑错，走回避重试算法
					if resp.GetCode() != 0 {
						atomic.AddUint64(&center.attempt, 1)
						break RECONNECT
					}
				}
			}
		}
	}
}

type TickerUpdate struct {
}

// 拉取最新tag到sdk
func (t TickerUpdate) update(ctx context.Context, center *ConfigCenter) {
	ticker := time.NewTicker(TickerUpdateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			info, err := center.pullVariables(ctx)
			if center.debug {
				log.Printf("getVariable: %+v, variables: %+v, cc: %+v, err: %+v", info, info.GetVariables(), center, err)
			}
			calCheckSum := checkSum(info.GetVariables())
			if err == nil && calCheckSum == info.GetCheckSum() && center.needUpdate(info) {
				center.updateVariables(info)
				if center.onChange != nil {
					center.onChange(center)
				}
				_ = center.backupVariables(info)
			} else if err == io.EOF || isRetryable(err, center.retryableCodes) {
				center.closeConn()
			}
		case <-ctx.Done():
			return
		}
	}
}
