package qudiscovery

import (
	"fmt"
	"net"
)

// GetLoaclIP 获取 eth 网卡的 ip
func GetLoaclIP(eth string) (string, error) {
	inters, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, inter := range inters {
		// 检查ip地址判断是否为 eth
		if inter.Name == eth {
			addrs, err := inter.Addrs()
			if err != nil {
				return "", err
			}
			if len(addrs) < 1 {
				return "", fmt.Errorf("can not found ip")
			}
			ip := addrs[0].(*net.IPNet)
			return ip.IP.String(), nil
		}
	}
	return "", fmt.Errorf("can not found %s", eth)
}
