package grpc

import (
	"context"
	"sync"
	"time"

	"git.qutoutiao.net/gopher/qms/internal/core/common"
	"git.qutoutiao.net/gopher/qms/internal/core/handler"
	"git.qutoutiao.net/gopher/qms/internal/core/invocation"
	"git.qutoutiao.net/gopher/qms/internal/core/server"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/iputil"
	"git.qutoutiao.net/gopher/qms/internal/pkg/util/protocol"
	"git.qutoutiao.net/gopher/qms/pkg/errors"
	"git.qutoutiao.net/gopher/qms/pkg/qenv"
	"git.qutoutiao.net/gopher/qms/pkg/qlog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

//err define
var (
	ErrGRPCSvcDescMissing = errors.New("must use server.WithRPCServiceDesc to set desc")
	ErrGRPCSvcType        = errors.New("must set *grpc.ServiceDesc")
)

//const
const (
	Name = "grpc"
)

func init() {
	server.InstallPlugin(Name, New)
}

//Server is grpc server holder
type grpcServer struct {
	s        *grpc.Server
	sonce    sync.Once // grpc.Server只注册一次
	opt      *serverOption
	handlers []handler.Handler // 暂时没有用到
}

//New create grpc server
func New(option *server.InitOptions) server.Server {
	return &grpcServer{
		opt: newServerOption(option),
	}
}

//Register register grpc services
func (s *grpcServer) Register(srv interface{}, opts ...server.Option) (string, error) {
	s.sonce.Do(func() {
		for _, o := range opts {
			o(s.opt)
		}

		// make interceptor chain
		s.opt.grpcOpts = append(s.opt.grpcOpts, grpc_middleware.WithUnaryServerChain(s.opt.unaryInts...))
		s.opt.grpcOpts = append(s.opt.grpcOpts, grpc_middleware.WithStreamServerChain(s.opt.streamInts...))

		// new server
		s.s = grpc.NewServer(s.opt.grpcOpts...)

		// Register reflection service on gRPC server.
		if s.opt.enableGrpcurl {
			reflection.Register(s.s)
		}
	})

	// 主要为了获取SvcDesc
	var opt serverOption
	for _, o := range opts {
		o(&opt)
	}

	if opt.svcDesc == nil {
		return "", ErrGRPCSvcDescMissing
	}

	// register service
	s.s.RegisterService(opt.svcDesc, srv)

	return "", nil
}

func (s *grpcServer) GetServer() interface{} {
	return s.s
}

//Start launch the server
func (s *grpcServer) Start() error {
	if s.s == nil { // 没有注册，直接返回
		return nil
	}

	listen := s.opt.listen
	if listen == nil {
		//l, host, port, lisErr := iputil.StartListener(s.opt.address, s.opt.tLSConfig)
		l, _, _, lisErr := iputil.StartListener(s.opt.address, s.opt.tLSConfig)
		if lisErr != nil {
			qlog.Error("listening failed, reason:" + lisErr.Error())
			return lisErr
		}
		//registry.InstanceEndpoints[s.opt.serverName] = net.JoinHostPort(host, port)
		listen = l
	}

	if laddr := listen.Addr().String(); !iputil.MatchServerPort(laddr, s.opt.address) {
		qlog.Panicf("服务端口不匹配，想要[%s]，实际[%s]", s.opt.address, laddr)
	}

	go func() {
		if err := s.s.Serve(listen); err != nil {
			qlog.Warn("grpc server err: " + err.Error())
			server.ErrRuntime <- err
		}
	}()

	qlog.Infof("%s server listening on: %s", s.opt.serverName, listen.Addr())
	return nil
}

//Stop gracfully shutdown grpc server
func (s *grpcServer) Stop() error {
	if s.s == nil { // 没有注册，直接返回
		return nil
	}

	stopped := make(chan struct{})
	go func() {
		s.s.GracefulStop()
		close(stopped)
	}()

	t := time.NewTimer(10 * time.Second)
	select {
	case <-t.C:
		s.s.Stop()
	case <-stopped:
		t.Stop()
	}
	return nil
}

//String return server name
func (s *grpcServer) String() string {
	return Name
}

//Request2Invocation convert grpc protocol to invocation
func Request2Invocation(ctx context.Context, serviceName string, req interface{}, info *grpc.UnaryServerInfo) *invocation.Invocation {
	md, _ := metadata.FromIncomingContext(ctx)
	sourceServices := md.Get(common.HeaderSourceName)
	var sourceService string
	if len(sourceServices) >= 1 {
		sourceService = sourceServices[0]
	}
	header := common.Header{}
	inv := &invocation.Invocation{
		MicroServiceName:   serviceName,
		SourceMicroService: sourceService,
		Args:               req,
		Protocol:           protocol.ProtocGrpc,
		Env:                qenv.Get(),
		SchemaID:           info.FullMethod,
		OperationID:        info.FullMethod,
		Ctx:                common.NewContext(ctx, header),
	}
	// set metadata to Ctx
	for k, vv := range md {
		header.Set(k, vv...)
	}
	return inv
}

//Stream2Invocation convert grpc protocol to invocation
func Stream2Invocation(serviceName string, stream grpc.ServerStream, info *grpc.StreamServerInfo) *invocation.Invocation {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	sourceServices := md.Get(common.HeaderSourceName)
	var sourceService string
	if len(sourceServices) >= 1 {
		sourceService = sourceServices[0]
	}
	header := common.Header{}
	inv := &invocation.Invocation{
		MicroServiceName:   serviceName,
		SourceMicroService: sourceService,
		Protocol:           protocol.ProtocGrpc,
		Env:                qenv.Get(),
		SchemaID:           info.FullMethod,
		OperationID:        info.FullMethod,
		Ctx:                common.NewContext(ctx, header),
	}
	// set metadata to Ctx
	for k, vv := range md {
		header.Set(k, vv...)
	}
	return inv
}
