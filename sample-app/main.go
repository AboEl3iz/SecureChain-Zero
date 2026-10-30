package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/securechain-zero/sample-app/internal/config"
	"github.com/securechain-zero/sample-app/internal/server"
)

func main() {
	// JSON structured logging for auditability and ingestion
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.LoadFromEnv()
	srv := server.New(cfg, logger)

	// Channel to listen for OS signals for graceful termination
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server fatal error", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdownSig:
		logger.Info("termination signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("server shutdown successfully completed")
	}
}
