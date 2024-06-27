package catalystgo

import (
	"fmt"
	"net"

	"github.com/catalystgo/logger/logger"
	"google.golang.org/grpc/reflection"
)

func (a *App) startGrpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.Grpc.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	reflection.Register(a.grpcServer)

	go func() {
		logger.Errorf(a.ctx, "gRPC server listening on port %d", a.cfg.Server.Grpc.Port)
		if err := a.grpcServer.Serve(lis); err != nil {
			logger.Fatalf(a.ctx, "failed to serve: %w", err)
		}
	}()

	a.grpcCloser.Add(func() error {
		logger.Errorf(a.ctx, "shutting down gRPC server")
		a.grpcServer.GracefulStop()
		return nil
	})

	return nil
}

func (a *App) startChannelz() error {
	return nil // TODO: Implement channelz
}
