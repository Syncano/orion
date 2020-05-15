package util

import (
	"context"

	"github.com/lithammer/shortuuid"
	"google.golang.org/grpc/metadata"
)

const medatadaRequestIDKey = "x-request-id"

func NewRequestID() string {
	return shortuuid.New()
}

func requestIDFromMetadata(md metadata.MD) string {
	header, ok := md[medatadaRequestIDKey]
	if !ok {
		return ""
	}

	reqID := header[0]

	return reqID
}

func RequestID(ctx context.Context, defaultReq func() string) string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		reqID := requestIDFromMetadata(md)
		if reqID != "" {
			return reqID
		}
	}

	md, ok = metadata.FromIncomingContext(ctx)
	if ok {
		reqID := requestIDFromMetadata(md)
		if reqID != "" {
			return reqID
		}
	}

	if defaultReq != nil {
		return defaultReq()
	}

	return ""
}

func DefaultRequestID(ctx context.Context) string {
	return RequestID(ctx, NewRequestID)
}

func AddRequestID(ctx context.Context, defaultReq func() string) (outCtx context.Context, reqID string) {
	reqID = RequestID(ctx, defaultReq)
	outCtx = metadata.AppendToOutgoingContext(ctx, medatadaRequestIDKey, reqID)

	return outCtx, reqID
}

func AddDefaultRequestID(ctx context.Context) (outCtx context.Context, reqID string) {
	return AddRequestID(ctx, NewRequestID)
}
