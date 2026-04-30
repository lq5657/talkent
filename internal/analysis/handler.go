package analysis

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/lq5657/talkent/internal/store"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/sessions/{id}/analyze", h.handleAnalyze)
	mux.HandleFunc("GET /api/sessions/{id}/report", h.handleGetReport)
	mux.HandleFunc("GET /api/sessions/{id}/reports", h.handleListReports)
}

type dimensionResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Score       int      `json:"score"`
	Comment     string   `json:"comment"`
	Suggestions []string `json:"suggestions"`
}

type analyzeResponse struct {
	ReportID  int64               `json:"report_id"`
	SessionID string              `json:"session_id"`
	Dimensions []dimensionResponse `json:"dimensions"`
	Markdown  string              `json:"markdown"`
	ModelUsed string              `json:"model_used"`
	CreatedAt string              `json:"created_at"`
}

func (h *Handler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	result, report, err := h.svc.TriggerAnalysis(r.Context(), sessionID, "manual")
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		case errors.Is(err, ErrSessionNotCompleted):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "session not completed, analysis only available for completed sessions"})
		default:
			h.logger.Error("analysis failed", "error", err, "session_id", sessionID)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to analyze session"})
		}
		return
	}

	dims := make([]dimensionResponse, len(result.DimensionResults))
	for i, d := range result.DimensionResults {
		dims[i] = dimensionResponse(d)
	}

	writeJSON(w, http.StatusCreated, analyzeResponse{
		ReportID:  report.ID,
		SessionID: sessionID,
		Dimensions: dims,
		Markdown:  result.Markdown,
		ModelUsed: result.ModelUsed,
		CreatedAt: report.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

type reportResponse struct {
	ReportID  int64               `json:"report_id"`
	SessionID string              `json:"session_id"`
	Dimensions []dimensionResponse `json:"dimensions"`
	Markdown  string              `json:"markdown"`
	ModelUsed string              `json:"model_used"`
	CreatedAt string              `json:"created_at"`
}

func (h *Handler) handleGetReport(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	report, err := h.svc.GetLatestReport(r.Context(), sessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "session not found"})
		default:
			h.logger.Error("get report failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get report"})
		}
		return
	}
	if report == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no report found for this session"})
		return
	}

	writeJSON(w, http.StatusOK, reportToResponse(report))
}

type reportSummary struct {
	ReportID  int64  `json:"report_id"`
	CreatedAt string `json:"created_at"`
	ModelUsed string `json:"model_used"`
}

type listReportsResponse struct {
	Reports []reportSummary `json:"reports"`
}

func (h *Handler) handleListReports(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	reports, err := h.svc.ListReports(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("list reports failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list reports"})
		return
	}

	summaries := make([]reportSummary, len(reports))
	for i, rp := range reports {
		summaries[i] = reportSummary{
			ReportID:  rp.ID,
			CreatedAt: rp.CreatedAt.Format("2006-01-02T15:04:05Z"),
			ModelUsed: rp.ModelUsed,
		}
	}

	writeJSON(w, http.StatusOK, listReportsResponse{Reports: summaries})
}

func reportToResponse(r *store.AnalysisReport) reportResponse {
	var dims []dimensionResponse
	if r.DimensionResults != "" {
		if err := json.Unmarshal([]byte(r.DimensionResults), &dims); err != nil {
			dims = nil
		}
	}
	return reportResponse{
		ReportID:  r.ID,
		SessionID: r.SessionID,
		Dimensions: dims,
		Markdown:  r.MarkdownContent,
		ModelUsed: r.ModelUsed,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
