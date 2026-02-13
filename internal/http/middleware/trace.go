package middleware

import (
	"net/http"
	"strings"

	"github.com/congregalis/aiden/pkg/traceid"
)

func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := strings.TrimSpace(r.Header.Get(traceid.HeaderName))
		if traceID == "" {
			traceID = traceid.Generate()
		}

		ctx := traceid.WithContext(r.Context(), traceID)
		w.Header().Set(traceid.HeaderName, traceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
