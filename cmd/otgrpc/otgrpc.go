package otgrpc

import (
	"context"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"google.golang.org/grpc/metadata"
)

// FilterFunc is a basic filter func for opentracing interceptor.
func FilterFunc(ctx context.Context, fullMethodName string) bool {
	spanCtx := opentracing.SpanFromContext(ctx)
	if spanCtx != nil {
		zipkinCtx, _ := spanCtx.Context().(zipkin.SpanContext)
		return zipkinCtx.Sampled
	}

	md, _ := metadata.FromIncomingContext(ctx)
	sampled := md["x-b3-sampled"]
	if sampled == nil {
		return false
	}
	ret, _ := strconv.ParseBool(sampled[0])
	return ret
}
