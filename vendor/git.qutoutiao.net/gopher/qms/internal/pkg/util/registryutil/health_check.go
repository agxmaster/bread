package registryutil

import (
	"encoding/json"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
)

const (
	DefaultHCInterval = 5 * time.Second
	DefaultHCTimeout  = 2 * time.Second
)

type HealthCheck struct {
	Interval string              `json:"interval"`
	Timeout  string              `json:"timeout"`
	HTTP     string              `json:"http"` // URI
	Header   map[string][]string `json:"header"`
	Method   string              `json:"method"`
	TCP      string              `json:"tcp"`
}

func (hc *HealthCheck) String() string {
	bytes, err := json.Marshal(hc)
	if err != nil {
		qlog.WithError(err).Errorf("marshal HealthCheck error")
		return ""
	}
	return string(bytes)
}

func Unmarshal(data []byte) *HealthCheck {
	var hc HealthCheck
	if err := json.Unmarshal(data, &hc); err != nil {
		qlog.WithError(err).Errorf("Unmarshal HealthCheck error")
		return nil
	}
	return &hc
}

func NewHealthCheck(appID, endpoint string) *HealthCheck {
	return &HealthCheck{
		Interval: DefaultHCInterval.String(),
		Timeout:  DefaultHCTimeout.String(),
		HTTP:     "http://" + endpoint + "/ping",
		Header: map[string][]string{
			constutil.ServiceHeader: {appID},
		},
		Method: "HEAD",
		TCP:    endpoint,
	}
}

func GetHealthCheck(meta map[string]string) *HealthCheck {
	if hcstr := meta["healthcheck"]; hcstr != "" {
		return Unmarshal([]byte(hcstr))
	}
	return nil
}
