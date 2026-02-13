package httpx

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/congregalis/aiden/internal/config"
	"github.com/congregalis/aiden/internal/http/handlers"
	"github.com/congregalis/aiden/internal/http/middleware"
)

func NewServer(cfg config.Config, logger *slog.Logger, readyFn handlers.ReadinessFunc) *http.Server {
	healthHandler := handlers.NewHealthHandler(readyFn)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("GET /readyz", healthHandler.Readyz)

	handler := middleware.TraceID(mux)
	handler = middleware.RequestLogger(logger, handler)

	return &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}
}
