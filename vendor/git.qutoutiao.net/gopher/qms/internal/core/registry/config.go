package registry

import (
	"strconv"
	"strings"
	"sync"

	qmsconfig "git.qutoutiao.net/gopher/qms/internal/config"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
)

var (
	config     *Config
	configOnce sync.Once
)

type Config struct {
	Consul   string                // consul地址
	Pilot    string                // pilot地址
	CacheDir string                // 缓存目录
	Local    map[string][]*Service // 本地列表
}

// GetConfig 指qms默认的配置
func GetConfig() (conf *Config, err error) {
	configOnce.Do(func() {
		rconf := qmsconfig.Get().Registry
		local := make(map[string][]*Service)

		for name, upstream := range qmsconfig.Get().Upstreams {
			if name == constutil.Common {
				continue
			}
			name = convertName(upstream.Address) // 调用的paasID
			for _, route := range upstream.CustomRoute {
				ipport := strings.Split(route.Address, ":")
				if len(ipport) != 2 {
					err = errors.Newf("service[%s] address[%s]格式应为IP:Port", name, route.Address)
					return
				}
				var port int
				port, err = strconv.Atoi(ipport[1])
				if err != nil {
					err = errors.WithStack(err)
					return
				}
				local[name] = append(local[name], &Service{
					Name:     name,
					IP:       ipport[0],
					Port:     port,
					Endpoint: ipport[0] + ":" + strconv.Itoa(port),
					Weight:   int32(route.Weight),
				})
			}
		}

		config = &Config{
			Consul:   rconf.Address,
			Pilot:    rconf.Pilot,
			CacheDir: rconf.CacheDir,
			Local:    local,
		}
	})
	return config, nil
}
