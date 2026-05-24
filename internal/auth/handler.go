package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lq5657/talkent/internal/config"
)

type Handler struct {
	jwtSvc *JWTService
	cfg    *config.AuthConfig
	logger *slog.Logger
}

func NewHandler(jwtSvc *JWTService, cfg *config.AuthConfig, logger *slog.Logger) *Handler {
	return &Handler{jwtSvc: jwtSvc, cfg: cfg, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/auth/refresh", h.handleRefresh)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username != h.cfg.Username || req.Password != h.cfg.Password {
		h.logger.Warn("login failed", "username", req.Username, "reason", "invalid credentials")
		writeAuthError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtSvc.GenerateAccessToken(req.Username)
	if err != nil {
		h.logger.Error("generate access token failed", "error", err)
		writeAuthError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtSvc.GenerateRefreshToken(req.Username)
	if err != nil {
		h.logger.Error("generate refresh token failed", "error", err)
		writeAuthError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("login successful", "username", req.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := h.jwtSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("refresh token validation failed", "reason", err.Error())
		writeAuthError(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtSvc.GenerateAccessToken(claims.Username)
	if err != nil {
		h.logger.Error("generate access token failed", "error", err)
		writeAuthError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refreshResponse{AccessToken: accessToken})
}
