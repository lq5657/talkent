package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const usernameKey contextKey = "username"

func UsernameFromContext(r *http.Request) string {
	if v, ok := r.Context().Value(usernameKey).(string); ok {
		return v
	}
	return ""
}

func AuthMiddleware(jwtSvc *JWTService, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip OPTIONS preflight
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Skip /health and /api/auth/*
			if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/api/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				writeAuthError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := jwtSvc.ValidateAccessToken(token)
			if err != nil {
				logger.Warn("auth failed", "reason", err.Error())
				writeAuthError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	// Authorization header (Bearer)
	if ah := r.Header.Get("Authorization"); ah != "" {
		if strings.HasPrefix(ah, "Bearer ") {
			return ah[7:]
		}
	}

	// Query param (for SSE/EventSource)
	if q := r.URL.Query().Get("token"); q != "" {
		return q
	}

	return ""
}

func writeAuthError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
