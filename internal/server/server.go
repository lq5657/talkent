package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lq5657/talkent/internal/config"
)

func New(cfg *config.Config, db *sql.DB, logger *slog.Logger, registerRoutes func(*http.ServeMux)) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler(db, logger))

	if registerRoutes != nil {
		registerRoutes(mux)
	}

	return &http.Server{
		Addr:    cfg.Addr(),
		Handler: mux,
	}
}

func healthHandler(db *sql.DB, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", "method", r.Method, "path", r.URL.Path)

		if err := db.Ping(); err != nil {
			logger.Error("health check failed: db ping", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","reason":"database ping failed"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func Run(srv *http.Server, logger *slog.Logger) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case sig := <-quit:
		logger.Info("shutting down", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	case err := <-errCh:
		return err
	}
}
