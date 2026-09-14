package routes

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/insanelyharsh/web-portfolio/internal/apperrors"
	"github.com/insanelyharsh/web-portfolio/internal/webserver/middlewares"
)

// writeError maps a service-layer error to an HTTP response shared by every
// route in this package, logging the real error (with the request's trace
// id, so it can be correlated with the access log line) before responding
// with a message safe to show a client. notFoundMessage lets each call site
// say what wasn't found (e.g. "blog not found" vs "comment not found").
func writeError(ctx context.Context, w http.ResponseWriter, err error, notFoundMessage string) {
	traceId := middlewares.TraceIdFromContext(ctx)

	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		slog.Warn(notFoundMessage, "trace_id", traceId, "error", err)
		http.Error(w, notFoundMessage, http.StatusNotFound)
	case errors.Is(err, apperrors.ErrValidation):
		slog.Warn("request validation failed", "trace_id", traceId, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error("internal server error", "trace_id", traceId, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
