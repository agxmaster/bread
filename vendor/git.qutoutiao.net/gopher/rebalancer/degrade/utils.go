package degrade

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

var tcpCheck = func(addr string, timeout time.Duration) error {
	// FIXME: 使用连接池，避免大量 TIME_WAIT
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func checkHealth(n *Node, defaultTimeout time.Duration) error {
	timeout := defaultTimeout

	if n.HealthCheck == nil {
		return tcpCheck(n.Host+":"+n.Port, timeout)
	}
	if n.HealthCheck.Timeout > 0 {
		timeout = n.HealthCheck.Timeout.Duration()
	}
	if len(n.HealthCheck.HTTP) == 0 {
		if len(n.HealthCheck.TCP) == 0 {
			return tcpCheck(n.Host+":"+n.Port, timeout)
		}
		return tcpCheck(n.HealthCheck.TCP, timeout)
	}

	method := n.HealthCheck.Method
	if len(method) == 0 {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, n.HealthCheck.HTTP, nil)
	if err != nil {
		return err
	}
	req.Header = n.HealthCheck.Header
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	req.Header.Set("User-Agent", "ReBalancer/0.0 static list check health")
	req.Header.Set("X-Qtt-Meshservice", n.Name)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil && err.Error() != io.EOF.Error() {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		return fmt.Errorf("http status:%d", resp.StatusCode)
	}
	if meshService := resp.Header.Get("X-Qtt-Mesh-Healthchecked-Service"); len(meshService) > 0 {
		if meshService != n.Name {
			return fmt.Errorf("response mesh service:%s is not :%s", meshService, n.Name)
		}
	}
	return nil
}

func getNormalServices(srvs []*Node) (normals []*Node, offlines map[string]*Node) {
	offlines = make(map[string]*Node)
	for i := range srvs {
		if srvs[i].IsOffline {
			offlines[srvs[i].Address] = srvs[i]
			continue
		}

		normals = append(normals, srvs[i])
	}
	return normals, offlines
}

// 严格比对两个服务器列表是否一致
func listSame(a, b []*Node) bool {
	if len(a) != len(b) {
		return false
	}

	checkA := make(map[string]bool)
	checkB := make(map[string]bool)
	for i := range a {
		checkA[a[i].GetSelfProtectionID()] = true
	}

	for i := range b {
		if !checkA[b[i].GetSelfProtectionID()] {
			return false
		}
		checkB[b[i].GetSelfProtectionID()] = true
	}

	for i := range a {
		if !checkB[a[i].GetSelfProtectionID()] {
			return false
		}
	}

	if len(checkA) != len(checkB) {
		return false
	}

	return true
}

func ctxIsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
