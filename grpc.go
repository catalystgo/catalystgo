package catalystgo

import (
	"github.com/catalystgo/logger/logger"
	"google.golang.org/grpc/reflection"
)

func (a *App) startGrpc() error {
	lis, err := newListener(a.cfg.Server.Grpc.Port)
	if err != nil {
		return err
	}

	go func() {
		logger.Errorf(a.ctx, "gRPC server listening on port %d", a.cfg.Server.Grpc.Port)

		err = a.grpcServer.Serve(lis)
		if err != nil {
			logger.Fatalf(a.ctx, "serve: %v", err)
		}
	}()
	reflection.Register(a.grpcServer)

	a.grpcCloser.Add(func() error {
		logger.Errorf(a.ctx, "shutdown gRPC server")
		a.grpcServer.GracefulStop()
		return nil
	})

	return nil
}

func (a *App) startChannelz() error {
	return nil // TODO: Implement channelz
}
