package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nseve/music-catalog/internal/core/db"
	"github.com/nseve/music-catalog/internal/core/logger"
	"github.com/nseve/music-catalog/internal/core/transport/middleware"
	"github.com/nseve/music-catalog/internal/core/transport/server"
	genre_repository "github.com/nseve/music-catalog/internal/features/genre/repository"
	genre_service "github.com/nseve/music-catalog/internal/features/genre/service"
	genre_transport "github.com/nseve/music-catalog/internal/features/genre/transport"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log, logFileClose, err := logger.New(logger.NewConfigMust())
	if err != nil {
		fmt.Println("create logger: %w", err)
		os.Exit(1)
	}
	defer func() {
		if err := logFileClose(); err != nil {
			fmt.Println("close log file:", err)
		}
	}()

	log.Info("initializing postgres connection pool")
	pool, err := db.New(ctx, db.NewConfigMust())
	if err != nil {
		log.Error("failed to create connection pool", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("initializing feature", slog.String("feature", "genre"))
	genreRepository := genre_repository.New(pool)
	genreService := genre_service.New(genreRepository)
	genreHandler := genre_transport.New(genreService)

	srv := server.New(
		server.NewConfigMust(),
		log,
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Panic(),
		middleware.Trace(),
	)

	v1 := server.NewAPIVersionRouter(server.ApiVersion1)
	v1.RegisterRoutes(genreHandler.Routes()...)
	srv.RegisterAPIRouters(v1)

	if err := srv.Run(ctx); err != nil {
		log.Error("server run error", slog.Any("error", err))
	}
}
