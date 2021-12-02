package metrics

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/qconf"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/json"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	"gopkg.in/yaml.v2"
)

var QmsBuckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 5, 10}
var restCustomLabels []CustomLabel

func enableErrorLogMetrics() error {
	return CreateCounter(&CounterOpts{
		Name:   ErrorNum,
		Help:   ErrorNumHelp,
		Labels: []string{ErrorNumLevel},
	})
}

func countErrorNumber(level string) {
	_ = CounterAdd(ErrorNum, 1, map[string]string{
		ErrorNumLevel: level,
	})
}

func enableProviderMetrics(customLabels ...CustomLabel) error {
	//REST
	labels := []string{QMSLabel, ReqProtocolLable, RespUriLable, RespCodeLable}
	if len(customLabels) > 5 {
		return errors.Newf("too many custom metric labels (%d)", len(customLabels))
	}
	for _, l := range customLabels {
		labels = append(labels, l.LabelName)
	}
	restCustomLabels = customLabels
	if err := CreateCounter(&CounterOpts{
		Name:   ReqQPS,
		Help:   ReqQPSHelp,
		Labels: labels,
	}); err != nil {
		return err
	}

	if err := CreateHistogram(&HistogramOpts{
		Name:    ReqDuration,
		Help:    ReqDurationHelp,
		Labels:  labels,
		Buckets: QmsBuckets,
	}); err != nil {
		return err
	}

	//GRPC
	if err := CreateCounter(&CounterOpts{
		Name:   GrpcReqQPS,
		Help:   GrpcReqQPSHelp,
		Labels: []string{QMSLabel, ReqProtocolLable, RespHandlerLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(&HistogramOpts{
		Name:    GrpcReqDuration,
		Help:    GrpcReqDurationHelp,
		Labels:  []string{QMSLabel, ReqProtocolLable, RespHandlerLable, RespCodeLable},
		Buckets: QmsBuckets,
	}); err != nil {
		return err
	}

	return nil
}

func GetRestCustomLabels() []CustomLabel {
	return restCustomLabels
}

func enableConsumerMetrics() error {
	//REST
	if err := CreateCounter(&CounterOpts{
		Name:   ClientReqQPS,
		Help:   ClientReqQPSHelp,
		Labels: []string{RemoteLable, ReqProtocolLable, RespUriLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(&HistogramOpts{
		Name:    ClientReqDuration,
		Help:    ClientReqDurationHelp,
		Labels:  []string{RemoteLable, ReqProtocolLable, RespUriLable, RespCodeLable},
		Buckets: QmsBuckets,
	}); err != nil {
		return err
	}

	//GRPC
	if err := CreateCounter(&CounterOpts{
		Name:   ClientGrpcReqQPS,
		Help:   ClientGrpcReqQPSHelp,
		Labels: []string{RemoteLable, ReqProtocolLable, RespHandlerLable, RespCodeLable},
	}); err != nil {
		return err
	}

	if err := CreateHistogram(&HistogramOpts{
		Name:    ClientGrpcReqDuration,
		Help:    ClientGrpcReqDurationHelp,
		Labels:  []string{RemoteLable, ReqProtocolLable, RespHandlerLable, RespCodeLable},
		Buckets: QmsBuckets,
	}); err != nil {
		return err
	}

	return nil
}

//1. only registry rest port(metrics on rest) and only for ces
func GetRegistryInstances() (instances []string) {
	ip := iputil.GetLocalIP()
	if len(ip) == 0 {
		qlog.Errorf("get localup failed")
		return
	}
	instance := ip + ":" + strconv.Itoa(config.Get().Native.Port)
	instances = append(instances, instance)

	return
}

type AutoMonitor struct{}

func isRegistryInstance() bool {
	instances := GetRegistryInstances()
	if instances == nil {
		qlog.Errorf("metircs registry instances not available")
		return false
	}
	Token := qconf.GetString("qms.metrics.autometrics.token", "80D2A4851C90AB1CC0842D55F409C518D") // key已废弃，兼容下
	Servicename := qconf.GetString("qms.metrics.autometrics.servicename", "app_exporter")          // key已废弃，兼容下
	Servicetype := qconf.GetString("qms.metrics.autometrics.servicetype", "prometheus_business")   // key已废弃，兼容下
	urlpath := config.Get().Metrics.AutoMetrics.Qurl

	url := urlpath + "?token=" + Token + "&servicetype=" + Servicetype + "&servicename=" + Servicename + "&instance=" + instances[0]
	cli := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		qlog.Errorf("new request failed, err:%s", err.Error())
		return false
	}
	resp, err := cli.Do(req)

	//resp, err := rest.ContextGet(context.TODO(), url)
	//if err != nil {
	//	if qerr.IsRateLimit(err) {
	//		qlog.Error("[rest]ratelimit: ", err)
	//	} else if qerr.IsCircuitBreak(err) {
	//		qlog.Error("[rest]circuitbreak: ", err)
	//	} else {
	//		qlog.Error("[rest]do request failed: ", err)
	//	}
	//	return false
	//}
	if err != nil {
		qlog.Errorf("request consul-api.qutoutiao.net failed, err:%s", err.Error())
		return false
	}
	if resp == nil || resp.Body == nil {
		qlog.Errorf("resp is nil")
		return false
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		qlog.Errorf("read resp body failed, err:%s", err.Error())
		return false
	}

	type QueryInfo struct {
		Requestid   string `json:"requestid"`
		Message     string `json:"message"`
		Servicetype string `json:"servicetype"`
		Instance    string `json:"instance"`
		Isexist     string `json:"isexist"`
	}

	var respInfo QueryInfo
	if err := yaml.Unmarshal(respBody, &respInfo); err != nil {
		qlog.Errorf("unmashal failed, err:%s", err.Error())
		return false
	}

	if respInfo.Isexist == "true" {
		qlog.Infof("has registry instance:%s  ", instances[0])
		return true
	}

	qlog.Infof("not registry instance:%s  ", instances[0])
	return false
}

func enableAutoRegistryMetrics() error {
	instances := GetRegistryInstances()
	if instances == nil {
		return errors.New("metircs registry instances not available")
	}

	type RegistryMetrics struct {
		Token        string   `json:"token"`
		Servicename  string   `json:"servicename"`
		Servicetype  string   `json:"servicetype"`
		Instancelist []string `json:"instancelist"`
		Tagnames     []string `json:"tagnames"`
	}

	service := config.Get().Service
	reqInfo := RegistryMetrics{
		Token:        qconf.GetString("qms.metrics.autometrics.token", "80D2A4851C90AB1CC0842D55F409C518D"), // key已废弃，兼容下
		Servicename:  qconf.GetString("qms.metrics.autometrics.servicename", "app_exporter"),                // key已废弃，兼容下
		Servicetype:  qconf.GetString("qms.metrics.autometrics.servicetype", "prometheus_business"),         // key已废弃，兼容下
		Instancelist: instances,
		Tagnames:     []string{"servicename=" + service.AppID, "serviceverion=" + service.Version},
	}

	reqData, err := json.Marshal(reqInfo)
	if err != nil {
		return errors.Wrap(err, "marshal reqInfo failed")
	}
	url := config.Get().Metrics.AutoMetrics.Url

	qlog.WithFields(qlog.Fields{
		"instances":       instances,
		"service_name":    service.AppID,
		"service_version": service.Version,
	}).Infof("注册metrics到运维consul[%s]上", url)

	cli := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "new request failed")
	}

	resp, err := cli.Do(req)
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		return errors.Wrap(err, "req post failed")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read resp body failed")
	}

	OutRespInfo(respBody, "register")
	return nil
}

func DeAutoRegistryMetrics() {
	instances := GetRegistryInstances()
	if instances == nil {
		qlog.Error("metircs deautoregistry, get instances not failed")
		return
	}

	type DeMetrics struct {
		Token        string   `json:"token"`
		Instancelist []string `json:"instancelist"`
	}

	reqInfo := DeMetrics{
		Token:        qconf.GetString("qms.metrics.autometrics.token", "80D2A4851C90AB1CC0842D55F409C518D"),
		Instancelist: instances,
	}

	reqData, err := json.Marshal(reqInfo)
	if err != nil {
		qlog.Errorf("Marshal reqInfo err:", err.Error())
		return
	}
	delurl := config.Get().Metrics.AutoMetrics.Deurl

	cli := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", delurl, bytes.NewBuffer(reqData))
	if err != nil {
		qlog.Errorf("new request failed, err:", err.Error())
		return
	}

	resp, err := cli.Do(req)
	if err != nil {
		qlog.Errorf("req post failed, err:", err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		qlog.Errorf("read resp body ", err.Error())
		return
	}

	OutRespInfo(respBody, "deleteipport")
}

func OutRespInfo(respBody []byte, action string) {
	type InstancsInfo struct {
		Status   string
		Msg      string
		Instance string
	}
	type RespInfo struct {
		Total       int
		Requestid   string
		Servicetype string
		Instances   []InstancsInfo
	}

	var respInfo RespInfo
	if err := yaml.Unmarshal(respBody, &respInfo); err != nil {
		qlog.Errorf("unmashal failed, err:%s", err.Error())
		return
	}
	if len(respInfo.Instances) != 0 {
		qlog.Infof("auto %s metrics info: status:%s  msg:%s ", action, respInfo.Instances[0].Status, respInfo.Instances[0].Msg)
	}

	//for _, in := range respInfo.Instances {
	//	qlog.Infof("auto %s metrics info: status:%s  msg:%s ", action, in.Status, in.Msg)
	//}
}

//Deprecated
func disableRedisMetrics() error {
	return nil
}

//Deprecated
func disableGormMetrics() error {
	return nil
}
