package mw

import (
	"context"

	"github.com/catalystgo/logger/logger"
	"google.golang.org/grpc"
)

func UnaryLoggingMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		logger.Errorf(ctx, err.Error())
		return nil, err
	}
	return resp, nil
}

func StreamLoggingMiddleware(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	err = handler(srv, stream)
	if err != nil {
		logger.Errorf(stream.Context(), err.Error())
		return err
	}
	return nil
}
