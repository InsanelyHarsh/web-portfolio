package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder wraps http.ResponseWriter to capture the status code that
// was actually written, since http.ResponseWriter doesn't expose it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// LoggerMiddleware logs one structured line per request (method, path,
// status, duration and trace id). It reads slog.Default() at request time
// rather than taking an injected logger, so it automatically picks up
// whatever logger.Init installs via slog.SetDefault at startup.
func LoggerMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		h.ServeHTTP(rec, r)

		slog.Default().Info("http request",
			"trace_id", TraceIdFromContext(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}
