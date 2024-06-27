package mw

import (
	"context"

	"github.com/catalystgo/catalystgo/errors"
	"google.golang.org/grpc"
)

func UnaryPanicMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.
				Newf("recovered from unary panic on %v", r).
				Code(errors.Internal).
				Op(info.FullMethod)
		}
	}()
	return handler(ctx, req)
}

func StreamPanicMiddleware(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.
				Newf("recovered from stream panic on %v", r).
				Code(errors.Internal).
				Op(info.FullMethod)
		}
	}()
	return handler(srv, stream)
}
