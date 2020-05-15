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

func RequestID(ctx context.Context, defaultReq func() string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return defaultReq()
	}

	header, ok := md[medatadaRequestIDKey]
	if !ok {
		return defaultReq()
	}

	reqID := header[0]
	if reqID == "" {
		return defaultReq()
	}

	return reqID
}

func AddRequestID(ctx context.Context, defaultReq func() string) context.Context {
	req := RequestID(ctx, defaultReq)

	return metadata.AppendToOutgoingContext(ctx, medatadaRequestIDKey, req)
}
