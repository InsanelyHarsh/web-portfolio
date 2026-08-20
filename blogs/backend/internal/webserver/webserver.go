package webserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/insanelyharsh/web-portfolio/internal/webserver/middlewares"
)

type WebServer struct {
	mux    *http.ServeMux
	server *http.Server
}

// Config holds everything InitWebServer needs to wire up the server. It
// exists so callers (main.go) build one env-driven value instead of passing
// a growing list of positional parameters.
type Config struct {
	Port               string
	CORSAllowedOrigins []string
	CORSAllowedHeaders []string
}

// InitWebServer wires up the mux with the standard middleware chain. Trace
// id is applied outermost so it populates the request context (and response
// header) before the logger middleware runs and reads it back for the log
// line. CORS sits innermost, directly around the mux, so preflight OPTIONS
// requests are still logged and trace-tagged even though CORS answers them
// itself instead of forwarding to a route.
func InitWebServer(cfg Config) *WebServer {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = middlewares.CORSMiddleware(cfg.CORSAllowedOrigins, cfg.CORSAllowedHeaders)(handler)
	handler = middlewares.LoggerMiddleware(handler)
	handler = middlewares.TraceIdMiddleware(handler)

	return &WebServer{
		mux: mux,
		server: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: handler,
		},
	}
}

// Mux exposes the underlying ServeMux so callers (main.go) can register
// routes without this package needing to know about them.
func (ws *WebServer) Mux() *http.ServeMux {
	return ws.mux
}

// Start blocks serving HTTP until the server is shut down or fails. A clean
// Shutdown is reported as a nil error, not http.ErrServerClosed.
func (ws *WebServer) Start() error {
	if err := ws.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server, waiting for in-flight requests to
// finish or ctx to be done, whichever comes first.
func (ws *WebServer) Shutdown(ctx context.Context) error {
	return ws.server.Shutdown(ctx)
}
