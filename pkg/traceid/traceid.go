package traceid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

const HeaderName = "X-Trace-Id"

type contextKey struct{}

// WithContext stores trace ID into context.
func WithContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKey{}, traceID)
}

// FromContext extracts trace ID from context.
func FromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(contextKey{}).(string); ok {
		return traceID
	}
	return ""
}

// Generate creates a random 16-byte trace ID.
func Generate() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "trace-id-unavailable"
	}
	return hex.EncodeToString(buf)
}
