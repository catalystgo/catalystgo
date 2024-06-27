package catalystgo

import (
	"context"
	"fmt"
	_ "net/http/pprof"
	"syscall"

	"github.com/catalystgo/catalystgo/closer"
	"github.com/catalystgo/healthcheck"
	"github.com/catalystgo/logger/logger"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type App struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	cfg *config

	hc healthcheck.Handler // TODO: Use healthcheck.Handler for liveness and readiness checks

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
	cfg, err := Parse()
	if err != nil {
		return nil, fmt.Errorf("parse config: %+v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		ctx:       ctx,
		ctxCancel: cancel,

		cfg: cfg,

		hc: healthcheck.NewHandler(),

		grpcServer:   grpc.NewServer(),
		publicServer: chi.NewMux(),
		adminServer:  chi.NewMux(),

		grpcCloser:  closer.New(),
		httpCloser:  closer.New(),
		adminCloser: closer.New(),
		globalCloser: closer.New(
			closer.WithSignals(syscall.SIGINT, syscall.SIGTERM),
			closer.WithTimeout(cfg.Server.GracefulShutdown.Timeout),
		),
	}

	app.globalCloser.Add(func() error {
		logger.Infof(ctx, "shutting down app")
		cancel()
		app.grpcCloser.CloseAll()
		app.httpCloser.CloseAll()
		app.adminCloser.CloseAll()
		return nil
	})

	return app, nil
}

func (a *App) Start(descriptions ...Service) error {
	// Get serviceDesc
	serviceDesc := make([]ServiceDesc, len(descriptions))
	for i, desc := range descriptions {
		serviceDesc[i] = desc.GetDescription()
	}

	a.desc = NewCompoundServiceDesc(serviceDesc...)

	a.desc.RegisterGRPC(a.grpcServer)
	a.desc.RegisterHTTP(a.ctx, a.newServerMuxHTTP())

	if err := a.startChannelz(); err != nil {
		return fmt.Errorf("start channelz: %+v", err)
	}
	if err := a.startGrpcServer(); err != nil {
		return fmt.Errorf("start grpc server: %+v", err)
	}
	if err := a.startHTTPServer(); err != nil {
		return fmt.Errorf("start http server: %+v", err)
	}
	if err := a.startAdminServer(); err != nil {
		return fmt.Errorf("start admin server: %+v", err)
	}

	return nil
}

func (a *App) Ctx() context.Context {
	return a.ctx
}
