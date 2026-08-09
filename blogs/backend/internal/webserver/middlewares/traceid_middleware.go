package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// ctxKey is an unexported type so trace-id values can't collide with context
// keys set by other packages (a raw string key would trip go vet's SA1029).
type ctxKey int

const traceIDKey ctxKey = iota

// TraceIdMiddleware assigns every request a random trace id, stores it on the
// request context, and echoes it back as a response header so callers can
// correlate a response with the corresponding server-side log lines.
func TraceIdMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceId := newTraceId()

		w.Header().Set("X-Trace-Id", traceId)

		ctx := context.WithValue(r.Context(), traceIDKey, traceId)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	}
}

// TraceIdFromContext returns the trace id set by TraceIdMiddleware, or "" if
// none is present (e.g. the context wasn't derived from a request that went
// through the middleware).
func TraceIdFromContext(ctx context.Context) string {
	traceId, _ := ctx.Value(traceIDKey).(string)
	return traceId
}

// newTraceId generates a random, non-cryptographic identifier for correlating
// a single request's logs. crypto/rand is used only for its convenient,
// dependency-free randomness source - the id has no security purpose.
func newTraceId() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}
