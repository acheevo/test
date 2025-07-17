package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/acheevo/test/internal/config"
	"github.com/acheevo/test/internal/http"
)

func main() {
	var cfg config.Config
	if err := cfg.Parse(); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
		}
	}()

	logger.Info("Starting Test API server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpServer, err := http.NewServer(logger, &cfg)
	if err != nil {
		logger.Fatal("Failed to create HTTP server", zap.Error(err))
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("Starting HTTP server", zap.String("addr", cfg.HTTPAddr))
		if err := httpServer.Start(ctx); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	case sig := <-sigCh:
		logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	logger.Info("Shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server", zap.Error(err))
	}

	logger.Info("Server shutdown complete")
}