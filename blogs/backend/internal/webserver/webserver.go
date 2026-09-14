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

type Config struct {
	Port               string
	CORSAllowedOrigins []string
	CORSAllowedHeaders []string
}

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

func (ws *WebServer) Mux() *http.ServeMux {
	return ws.mux
}

func (ws *WebServer) Start() error {
	if err := ws.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (ws *WebServer) Shutdown(ctx context.Context) error {
	return ws.server.Shutdown(ctx)
}
