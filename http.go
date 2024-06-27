package catalystgo

import (
	"fmt"
	"net"
	"net/http"

	"github.com/catalystgo/logger/logger"
	"github.com/flowchartsman/swaggerui"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func (a *App) newServerMuxHTTP() *runtime.ServeMux {
	mux := runtime.NewServeMux()
	a.publicServer.Mount("/", mux)
	return mux
}

func (a *App) startHTTPServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.HTTP.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	publicServer := http.Server{Handler: a.publicServer}
	go func() {
		logger.Infof(a.ctx, "HTTP server listening on port %d", a.cfg.Server.HTTP.Port)
		if err := http.Serve(lis, a.publicServer); err != nil {
			logger.Fatalf(a.ctx, "serve: %v", err)
		}
	}()

	a.httpCloser.Add(func() error {
		logger.Infof(a.ctx, "shutting down HTTP server")
		return publicServer.Shutdown(a.ctx)
	})

	return nil
}

func (a *App) startAdminServer() error {
	a.setupAdminServer()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.Server.Debug.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	adminServer := http.Server{Handler: a.adminServer}
	go func() {
		logger.Infof(a.ctx, "Admin server listening on port %d", a.cfg.Server.Debug.Port)
		if err := http.Serve(lis, a.adminServer); err != nil {
			logger.Fatalf(a.ctx, "serve: %v", err)
		}
	}()

	a.httpCloser.Add(func() error {
		logger.Infof(a.ctx, "shutting down Admin server")
		return adminServer.Shutdown(a.ctx)
	})

	return nil
}

func (a *App) setupAdminServer() {
	b := a.desc.SwaggerJSON()

	a.adminServer.Handle("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		w.Write(b)
	}))
	a.adminServer.Handle("/docs/", http.StripPrefix("/docs/", swaggerui.Handler(b)))
	a.adminServer.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: Use healthcheck handler
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	a.adminServer.Handle("/readyz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // TODO: Use healthcheck handler
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
}
