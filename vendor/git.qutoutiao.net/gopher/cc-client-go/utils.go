package cc

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-ps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrEmptyIP an error for empty ip strings
	ErrEmptyIP = errors.New("empty string given for ip")

	// ErrNotHostColonPort an error for invalid host port string
	ErrNotHostColonPort = errors.New("expecting host:port")

	// ErrNotFourOctets an error for the wrong number of octets after splitting a string
	ErrNotFourOctets = errors.New("Wrong number of octets")
)

// ParseIPToUint32 converts a string ip (e.g. "x.y.z.w") to an uint32
func ParseIPToUint32(ip string) (uint32, error) {
	if ip == "" {
		return 0, ErrEmptyIP
	}

	if ip == "localhost" {
		return 127<<24 | 1, nil
	}

	octets := strings.Split(ip, ".")
	if len(octets) != 4 {
		return 0, ErrNotFourOctets
	}

	var intIP uint32
	for i := 0; i < 4; i++ {
		octet, err := strconv.Atoi(octets[i])
		if err != nil {
			return 0, err
		}
		intIP = (intIP << 8) | uint32(octet)
	}

	return intIP, nil
}

// ParsePort converts port number from string to uin16
func ParsePort(portString string) (uint16, error) {
	port, err := strconv.ParseUint(portString, 10, 16)
	return uint16(port), err
}

// PackIPAsUint32 packs an IPv4 as uint32
func PackIPAsUint32(ip net.IP) uint32 {
	if ipv4 := ip.To4(); ipv4 != nil {
		return binary.BigEndian.Uint32(ipv4)
	}
	return 0
}

// This code is borrowed from https://github.com/uber/tchannel-go/blob/dev/localip.go

// scoreAddr scores how likely the given addr is to be a remote address and returns the
// IP to use when listening. Any address which receives a negative score should not be used.
// Scores are calculated as:
// -1 for any unknown IP addresses.
// +300 for IPv4 addresses
// +100 for non-local addresses, extra +100 for "up" interaces.
func scoreAddr(iface net.Interface, addr net.Addr) (int, net.IP) {
	var ip net.IP
	if netAddr, ok := addr.(*net.IPNet); ok {
		ip = netAddr.IP
	} else if netIP, ok := addr.(*net.IPAddr); ok {
		ip = netIP.IP
	} else {
		return -1, nil
	}

	var score int
	if ip.To4() != nil {
		score += 300
	}
	if iface.Flags&net.FlagLoopback == 0 && !ip.IsLoopback() {
		score += 100
		if iface.Flags&net.FlagUp != 0 {
			score += 100
		}
	}
	return score, ip
}

// HostIP tries to find an IP that can be used by other machines to reach this machine.
func HostIP() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	bestScore := -1
	var bestIP net.IP
	// Select the highest scoring IP as the best IP.
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			// Skip this interface if there is an error.
			continue
		}

		for _, addr := range addrs {
			score, ip := scoreAddr(iface, addr)
			if score > bestScore {
				bestScore = score
				bestIP = ip
			}
		}
	}

	if bestScore == -1 {
		return nil, errors.New("no addresses to listen on")
	}

	return bestIP, nil
}

// CurrentProcessName 当前进程的名称
func CurrentProcessName() (string, error) {
	pid := os.Getpid()
	proc, err := ps.FindProcess(pid)
	if err != nil {
		return "", err
	}
	return proc.Executable(), nil
}

func MD5CheckSum(s string) string {
	h := md5.New()
	_, err := io.WriteString(h, s)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func isRetryable(err error, retryableCodes []codes.Code) bool {
	errCode := status.Code(err)
	if isContextError(err) {
		return false
	}
	for _, code := range retryableCodes {
		if code == errCode {
			return true
		}
	}
	return false
}

func isContextError(err error) bool {
	switch status.Code(err) {
	case codes.DeadlineExceeded, codes.Canceled:
		return true
	default:
		return false
	}
}

func DoWithTimeout(f func(center *ConfigCenter) error, d time.Duration, center *ConfigCenter) error {
	errChan := make(chan error, 1)
	go func(center *ConfigCenter) {
		errChan <- f(center)
		close(errChan)
	}(center)
	t := time.NewTimer(d)
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

func DoWithTimeoutClosure(f func(*ConfigCenter) error, d time.Duration) func(*ConfigCenter) error {
	return func(center *ConfigCenter) error {
		return DoWithTimeout(f, d, center)
	}
}
