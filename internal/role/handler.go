package role

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/roles/recommend-goals", h.handleRecommendGoals)
	mux.HandleFunc("POST /api/roles/recommend-dimensions", h.handleRecommendDimensions)
}

type recommendGoalsRequest struct {
	RoleDescription string `json:"role_description"`
}

type recommendGoalsResponse struct {
	Source string          `json:"source"` // "template" or "llm"
	Goals  []TrainingGoal  `json:"goals"`
}

func (h *Handler) handleRecommendGoals(w http.ResponseWriter, r *http.Request) {
	var req recommendGoalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.RoleDescription == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_description is required"})
		return
	}

	goals, err := h.svc.RecommendGoals(r.Context(), req.RoleDescription)
	if err != nil {
		h.logger.Error("recommend goals failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to recommend goals"})
		return
	}

	source := "template"
	if _, ok := MatchTemplate(req.RoleDescription); !ok {
		source = "llm"
	}

	writeJSON(w, http.StatusOK, recommendGoalsResponse{Source: source, Goals: goals})
}

type recommendDimensionsRequest struct {
	RoleType  RoleType        `json:"role_type"`
	Goals     []TrainingGoal  `json:"goals"`
	Mode      string          `json:"mode"`      // "table" (default) or "derive"
	RoleDesc  string          `json:"role_desc"` // required when mode=derive
}

type recommendDimensionsResponse struct {
	Source     string       `json:"source"` // "table" or "llm"
	Dimensions []Dimension  `json:"dimensions"`
}

func (h *Handler) handleRecommendDimensions(w http.ResponseWriter, r *http.Request) {
	var req recommendDimensionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Mode == "derive" {
		if req.RoleDesc == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_desc is required when mode=derive"})
			return
		}

		dims, err := h.svc.DeriveDimensions(r.Context(), req.RoleDesc, req.Goals)
		if err != nil {
			h.logger.Error("derive dimensions failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to derive dimensions"})
			return
		}

		writeJSON(w, http.StatusOK, recommendDimensionsResponse{Source: "llm", Dimensions: dims})
		return
	}

	if req.RoleType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_type is required"})
		return
	}

	dims, err := h.svc.RecommendDimensions(r.Context(), req.RoleType, req.Goals)
	if err != nil {
		h.logger.Error("recommend dimensions failed", "error", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, recommendDimensionsResponse{Source: "table", Dimensions: dims})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
