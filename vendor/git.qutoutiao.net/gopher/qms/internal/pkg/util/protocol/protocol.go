package protocol

import (
	"strings"

	"git.qutoutiao.net/gopher/qms/internal/pkg/util/constutil"
)

type Protocol int

const (
	ProtocUnknown Protocol = iota
	ProtocHTTP
	ProtocGrpc
)

const rest = "rest"

func (p Protocol) IsValid() bool {
	switch p {
	case ProtocHTTP, ProtocGrpc:
		return true
	}
	return false
}

func (p Protocol) String() string {
	switch p {
	case ProtocHTTP:
		return constutil.HTTP
	case ProtocGrpc:
		return constutil.GRPC
	default:
		return constutil.Unknown
	}
}

func ToProtocol(p string) Protocol {
	switch strings.ToLower(p) {
	case constutil.HTTP, rest:
		return ProtocHTTP
	case constutil.GRPC:
		return ProtocGrpc
	default:
		return ProtocUnknown
	}
}
