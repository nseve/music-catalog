package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nseve/music-catalog/internal/core/transport/middleware"
)

type Server struct {
	mux         *http.ServeMux
	config      Config
	log         *slog.Logger
	middlewares []middleware.Middleware
}

func New(config Config, log *slog.Logger, mws ...middleware.Middleware) *Server {
	return &Server{
		mux:         http.NewServeMux(),
		config:      config,
		log:         log,
		middlewares: mws,
	}
}

func (h *Server) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.apiVersion)
		h.mux.Handle(prefix+"/", http.StripPrefix(prefix, router))
	}
}

func (h *Server) Run(ctx context.Context) error {
	mux := middleware.ChainMiddleware(h.mux, h.middlewares...)

	server := &http.Server{
		Addr:    h.config.Addr,
		Handler: mux,
	}

	ch := make(chan error, 1)

	go func() {
		defer close(ch)
		h.log.Info("start HTTP server", slog.String("addr", h.config.Addr))
		ch <- server.ListenAndServe()
	}()

	select {
	case err := <-ch:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
	case <-ctx.Done():
		h.log.Info("shutting down HTTP server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			h.config.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		h.log.Info("HTTP server stopped")
	}

	return nil
}
