package cc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewGRPCConn(ctx context.Context, env Env, qaServerURL string) (*grpc.ClientConn, error) {
	cp := x509.NewCertPool()
	var addr string
	switch env {
	case QA, PG, DEV:
		addr = qaServerURL
		cp.AppendCertsFromPEM([]byte(QACrt))
	default:
		addr = PRDServerAddr
		cp.AppendCertsFromPEM([]byte(PrdCrt))
	}
	creds := credentials.NewTLS(&tls.Config{RootCAs: cp, ServerName: "server.grpc.io"})
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(0),
	}
	// stream 手动重试
	streamRetryOpts := []grpc_retry.CallOption{
		grpc_retry.WithMax(0),
	}
	// 手动重试两次
	var conn *grpc.ClientConn
	var err error
	for i := 0; i < 2; i++ {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithDisableRetry(), grpc.WithTransportCredentials(creds),
			grpc.WithWriteBufferSize(0),
			grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(streamRetryOpts...)),
			grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)))
		if err == nil {
			return conn, err
		}
		time.Sleep(200 * time.Millisecond)
	}
	// conn不返回nil
	return &grpc.ClientConn{}, err
}
