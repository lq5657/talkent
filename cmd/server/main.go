package main

import (
	"flag"
	"log/slog"
	"net/http"
	"context"
	"os"

	"github.com/lq5657/talkent/internal/auth"
	"github.com/lq5657/talkent/internal/config"
	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/log"
	"github.com/lq5657/talkent/internal/memory"
	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/server"
	"github.com/lq5657/talkent/internal/session"
	"github.com/lq5657/talkent/internal/analysis"
	"github.com/lq5657/talkent/internal/store"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).
			Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := log.New(&cfg.Log)
	logger.Info("config loaded", "path", *configPath)

	db, err := store.Open(cfg.Database.Path)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database initialized", "path", cfg.Database.Path)

	llmClient, err := llm.NewClient(&cfg.LLM, logger)
	if err != nil {
		logger.Error("failed to initialize LLM client", "error", err)
		os.Exit(1)
	}
	logger.Info("LLM client initialized", "provider", cfg.LLM.Provider, "model", cfg.LLM.Model)

	roleSvc := role.NewService(llmClient, logger)
	roleHandler := role.NewHandler(roleSvc, logger)

	sessionStore := store.NewSessionStore(db)
	memoryManager := memory.NewManager(cfg.Session.MemoryWindowSize, llmClient, logger)
	sessionSvc := session.NewService(sessionStore, memoryManager, llmClient, logger)
	sessionHandler := session.NewHandler(sessionSvc, logger)

	analysisStore := store.NewAnalysisStore(db)
	analysisEngine := analysis.NewEngine(llmClient, logger)
	analysisSvc := analysis.NewService(analysisStore, analysisEngine, sessionStore, logger)
	analysisHandler := analysis.NewHandler(analysisSvc, logger)

	if cfg.Analysis.AutoTrigger {
		sessionSvc.OnSessionEnd = func(ctx context.Context, sessionID string) {
			if _, _, err := analysisSvc.TriggerAnalysis(ctx, sessionID, "auto"); err != nil {
				logger.Warn("auto analysis failed", "session_id", sessionID, "error", err)
			}
		}
	}

	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.AccessExpiry, cfg.Auth.RefreshExpiry)
	authHandler := auth.NewHandler(jwtSvc, &cfg.Auth, logger)
	authMw := auth.AuthMiddleware(jwtSvc, logger)

	srv := server.New(cfg, db, logger, func(mux *http.ServeMux) {
		authHandler.RegisterRoutes(mux)
		roleHandler.RegisterRoutes(mux)
		sessionHandler.RegisterRoutes(mux)
		analysisHandler.RegisterRoutes(mux)

		// Frontend static files + SPA fallback
		mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/dist/assets"))))
		mux.Handle("/", server.SpaHandler(http.Dir("web/dist")))
	}, authMw)

	if err := server.Run(srv, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
