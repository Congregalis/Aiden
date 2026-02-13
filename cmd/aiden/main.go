package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/congregalis/aiden/internal/config"
	httpx "github.com/congregalis/aiden/internal/http"
	"github.com/congregalis/aiden/internal/logger"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		slog.Error("load config failed", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger.New(cfg.Log)
	log.Info("configuration loaded",
		slog.String("env", cfg.AppEnv),
		slog.String("http_port", cfg.HTTP.Port),
	)

	server := httpx.NewServer(cfg, log, nil)

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", slog.String("addr", server.Addr))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		if err != nil {
			log.Error("server exited unexpectedly", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	select {
	case err := <-errCh:
		if err != nil {
			log.Error("server stopped with error", slog.Any("error", err))
			os.Exit(1)
		}
	case <-time.After(100 * time.Millisecond):
	}

	log.Info("server stopped")
}
