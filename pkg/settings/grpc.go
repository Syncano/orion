package settings

import (
	"time"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/Syncano/orion/cmd/otgrpc"
)

const (
	// MaxGRPCMessageSize defines max send/recv grpc payload.
	MaxGRPCMessageSize = 10 << 20
	// KeepaliveParamsTime defines duration after which server will ping client to see if the transport is still alive.
	KeepaliveParamsTime = 10 * time.Second
	// KeepaliveParamsTimeout defines duration that server waits for client to respond.
	KeepaliveParamsTimeout = 5 * time.Second
)

var (
	// DefaultGRPCServerOptions defines default grpc server options (duh).
	DefaultGRPCServerOptions = []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithFilterFunc(otgrpc.FilterFunc)),
		),
		grpc.StreamInterceptor(
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithFilterFunc(otgrpc.FilterFunc)),
		),
		grpc.MaxRecvMsgSize(MaxGRPCMessageSize),
		grpc.MaxSendMsgSize(MaxGRPCMessageSize),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    KeepaliveParamsTime,
			Timeout: KeepaliveParamsTimeout,
		}),
	}

	// DefaultGRPCDialOptions defines default grpc dial options.
	DefaultGRPCDialOptions = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxGRPCMessageSize)),
		grpc.WithUnaryInterceptor(
			grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithFilterFunc(otgrpc.FilterFunc)),
		),
		grpc.WithStreamInterceptor(
			grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithFilterFunc(otgrpc.FilterFunc)),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    KeepaliveParamsTime,
			Timeout: KeepaliveParamsTimeout,
		}),
	}
)
