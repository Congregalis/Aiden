package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/congregalis/aiden/internal/config"
	"github.com/congregalis/aiden/internal/db"
	httpx "github.com/congregalis/aiden/internal/http"
	"github.com/congregalis/aiden/internal/logger"
	"github.com/congregalis/aiden/internal/telegram"
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

	dbConn, err := db.Open(cfg.Database)
	if err != nil {
		log.Error("open database failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer dbConn.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := dbConn.PingContext(pingCtx); err != nil {
		log.Error("database ping failed", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("database connected")

	readinessFn := func(ctx context.Context) error {
		return dbConn.PingContext(ctx)
	}

	server := httpx.NewServer(cfg, log, readinessFn)

	rootCtx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	telegramClient := telegram.NewHTTPClient(cfg.Telegram.BotToken, nil)
	telegramStore := telegram.NewSQLStore(dbConn)
	telegramWorker := telegram.NewWorker(telegram.WorkerConfig{
		PollTimeoutSec: cfg.Telegram.PollTimeoutSec,
		PollInterval:   time.Duration(cfg.Telegram.PollIntervalMS) * time.Millisecond,
		AllowedUpdates: telegram.ParseAllowedUpdates(cfg.Telegram.AllowedUpdates),
	}, telegramClient, telegramStore, log)

	serverErrCh := make(chan error, 1)
	botErrCh := make(chan error, 1)

	var workerWG sync.WaitGroup
	workerWG.Add(1)
	go func() {
		defer workerWG.Done()
		if err := telegramWorker.Run(rootCtx); err != nil {
			botErrCh <- err
			return
		}
		botErrCh <- nil
	}()

	go func() {
		log.Info("http server listening", slog.String("addr", server.Addr))
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	exitCode := 0
	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-serverErrCh:
		if err != nil {
			log.Error("server exited unexpectedly", slog.Any("error", err))
			exitCode = 1
		}
	case err := <-botErrCh:
		if err != nil {
			log.Error("telegram worker exited unexpectedly", slog.Any("error", err))
			exitCode = 1
		}
	}

	cancelRoot()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server shutdown failed", slog.Any("error", err))
		exitCode = 1
	}

	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server stopped with error", slog.Any("error", err))
			exitCode = 1
		}
	case <-time.After(100 * time.Millisecond):
	}

	done := make(chan struct{})
	go func() {
		workerWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(cfg.HTTP.ShutdownTimeout):
		log.Warn("telegram worker shutdown timeout reached")
		exitCode = 1
	}

	log.Info("server stopped")

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
