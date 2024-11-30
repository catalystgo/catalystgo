package catalystgo

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"syscall"
	"time"

	"github.com/catalystgo/catalystgo/closer"
	"github.com/catalystgo/catalystgo/internal/config"
	"github.com/catalystgo/logger/logger"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type App struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	cfg *config.AppConfig

	desc ServiceDesc

	grpcServer   *grpc.Server
	publicServer chi.Router
	adminServer  chi.Router

	grpcCloser   closer.Closer
	httpCloser   closer.Closer
	adminCloser  closer.Closer
	globalCloser closer.Closer
}

func New() (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Parse(ctx, *configPath)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("parse config: %+v", err)
	}

	app := &App{
		ctx:       ctx,
		ctxCancel: cancel,

		cfg: cfg,

		grpcServer:   grpc.NewServer(),
		publicServer: chi.NewMux(),
		adminServer:  chi.NewMux(),

		grpcCloser:  closer.New(),
		httpCloser:  closer.New(),
		adminCloser: closer.New(),

		globalCloser: closer.New(
			closer.WithSignals(syscall.SIGINT, syscall.SIGTERM),
			closer.WithTimeout(cfg.Server.Shutdown.Timeout),
		),
	}

	app.globalCloser.Add(func() error {
		logger.ErrorKV(ctx, "got termination signal")

		logger.ErrorKV(ctx, "shutting down app")

		app.grpcCloser.CloseAll()
		app.httpCloser.CloseAll()
		app.adminCloser.CloseAll()
		logger.ErrorKV(ctx, "traffic closed")

		logger.Errorf(ctx, "shutting down in %s", cfg.Server.Shutdown.Delay.String())
		time.Sleep(cfg.Server.Shutdown.Delay)

		closer.CloseAll()

		cancel()
		return nil
	})

	return app, nil
}

func (a *App) Run(descriptions ...Service) error {
	serviceDesc := make([]ServiceDesc, len(descriptions))
	for i, desc := range descriptions {
		serviceDesc[i] = desc.GetDescription()
	}

	a.desc = newCompoundServiceDesc(serviceDesc...)

	a.desc.RegisterGRPC(a.grpcServer)
	err := a.desc.RegisterHTTP(a.ctx, a.newServerMuxHTTP())
	if err != nil {
		return err
	}

	if err = a.startChannelz(); err != nil {
		return errors.Errorf("start channelz: %v", err)
	}
	if err = a.startGrpc(); err != nil {
		return errors.Errorf("start grpc server: %v", err)
	}
	if err = a.startHTTP(); err != nil {
		return errors.Errorf("start http server: %v", err)
	}
	if err = a.startAdmin(); err != nil {
		return errors.Errorf("start admin server: %v", err)
	}

	logger.Error(a.ctx, "app running")

	<-a.ctx.Done() // Wait for the app to be closed

	return nil
}

func (a *App) Ctx() context.Context {
	return a.ctx
}
