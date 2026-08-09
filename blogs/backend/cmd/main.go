package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/insanelyharsh/web-portfolio/internal/blog"
	"github.com/insanelyharsh/web-portfolio/internal/blog/repository"
	"github.com/insanelyharsh/web-portfolio/internal/config"
	"github.com/insanelyharsh/web-portfolio/internal/logger"
	"github.com/insanelyharsh/web-portfolio/internal/migration"
	"github.com/insanelyharsh/web-portfolio/internal/webserver"
	"github.com/insanelyharsh/web-portfolio/internal/webserver/routes"
)

func main() {
	// DATABASE_URL, PORT, and LOG_LEVEL are expected to be set externally —
	// via the shell for local `go run`, or via docker-compose's env_file for
	// the containerized setup (see docker-compose.yml).
	logger.Init()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := config.InitPostgres(ctx)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if err := migration.Run(ctx, conn); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// NewBlogRepository takes pgx.Conn by value; a single one-time copy off
	// the pointer InitPostgres returns is safe here since nothing else
	// queries through the original *pgx.Conn afterward (only Close, below).
	repo := repository.NewBlogRepository(*conn)
	manager := blog.NewBlogManager(repo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ws := webserver.InitWebServer(port)
	routes.RegisterBlogRoutes(ws.Mux(), manager)

	serveErr := make(chan error, 1)
	go func() {
		slog.Info("starting webserver", "port", port)
		serveErr <- ws.Start()
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			slog.Error("webserver stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		slog.Info("shutting down webserver")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := ws.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	}
}
