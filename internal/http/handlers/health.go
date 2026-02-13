package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/congregalis/aiden/pkg/traceid"
)

type ReadinessFunc func(context.Context) error

type HealthHandler struct {
	startedAt time.Time
	readyFn   ReadinessFunc
}

func NewHealthHandler(readyFn ReadinessFunc) HealthHandler {
	return HealthHandler{
		startedAt: time.Now().UTC(),
		readyFn:   readyFn,
	}
}

func (h HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":   "ok",
		"service":  "aiden",
		"trace_id": traceid.FromContext(r.Context()),
		"uptime":   time.Since(h.startedAt).String(),
	}
	writeJSON(w, http.StatusOK, response)
}

func (h HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.readyFn != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.readyFn(ctx); err != nil {
			response := map[string]any{
				"status":   "not_ready",
				"trace_id": traceid.FromContext(r.Context()),
			}
			writeJSON(w, http.StatusServiceUnavailable, response)
			return
		}
	}

	response := map[string]any{
		"status":   "ready",
		"trace_id": traceid.FromContext(r.Context()),
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
