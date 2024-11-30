package catalystgo

import (
	"net/http"

	"github.com/catalystgo/logger/logger"
	"github.com/flowchartsman/swaggerui"
	"github.com/go-chi/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
)

func (a *App) newServerMuxHTTP() *runtime.ServeMux {
	mux := runtime.NewServeMux()
	a.publicServer.Mount("/", mux)
	return mux
}

func (a *App) setupHTTP() {
	// Basic CORS, for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	a.publicServer.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		//AllowedOrigins: []string{"https://*", "http://*"},
		//AllowedOrigins: []string{"*"},
		AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
}

func (a *App) startHTTP() error {
	lis, err := newListener(a.cfg.Server.HTTP.Port)
	if err != nil {
		return err
	}

	publicServer := http.Server{Handler: a.publicServer}

	go func() {
		logger.Errorf(a.ctx, "http server listening on port %d", a.cfg.Server.HTTP.Port)

		err = http.Serve(lis, a.publicServer)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf(a.ctx, "serve: %v", err)
		}
	}()

	a.httpCloser.Add(func() error {
		logger.Errorf(a.ctx, "shutdown http server")
		return publicServer.Shutdown(a.ctx)
	})

	return nil
}

func (a *App) startAdmin() error {
	a.setupAdminServer()

	lis, err := newListener(a.cfg.Server.Admin.Port)

	adminServer := http.Server{Handler: a.adminServer}
	go func() {
		logger.Errorf(a.ctx, "admin server listening on port %d", a.cfg.Server.Admin.Port)

		err = http.Serve(lis, a.adminServer)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf(a.ctx, "serve admin: %v", err)
		}
	}()

	a.httpCloser.Add(func() error {
		logger.Errorf(a.ctx, "shutdown admin server")
		return adminServer.Shutdown(a.ctx)
	})

	return nil
}

func (a *App) setupAdminServer() {
	b := a.desc.SwaggerJSON()

	// Swagger UI

	a.adminServer.Handle("/swagger.json", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	a.adminServer.Handle("/docs/", http.StripPrefix("/docs/", swaggerui.Handler(b)))

	// Healthcheck

	a.adminServer.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO: Use healthcheck handler
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	a.adminServer.Handle("/ready", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// TODO: Use healthcheck handler
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
}
