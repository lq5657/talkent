package analysis

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/lq5657/talkent/internal/store"
)

func TestHandler_Analyze_Success(t *testing.T) {
	// V8: POST /api/sessions/{id}/analyze returns 201
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "analyze-ok",
		RoleConfig: `{"description":"test","scenario":"test"}`,
		Goals:      "[]",
		Dimensions: `[{"name":"test","description":"d"}]`,
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	engine := &mockEngine{
		result: &AnalysisResult{
			DimensionResults: []DimensionResult{{Name: "test", Description: "d", Score: 8, Comment: "good", Suggestions: []string{"improve"}}},
			Markdown:         "# Report",
			ModelUsed:        "test-model",
		},
	}
	svc := NewService(analysisStore, engine, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/api/sessions/analyze-ok/analyze", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp analyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.SessionID != "analyze-ok" {
		t.Errorf("session_id = %q, want %q", resp.SessionID, "analyze-ok")
	}
	if resp.ReportID == 0 {
		t.Error("report_id should be set")
	}
	if len(resp.Dimensions) != 1 {
		t.Errorf("dimensions = %d, want 1", len(resp.Dimensions))
	}
}

func TestHandler_Analyze_ActiveSession(t *testing.T) {
	// V7 (via handler): active session returns 409
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "active-session",
		RoleConfig: `{"description":"test"}`,
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, &mockEngine{}, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/api/sessions/active-session/analyze", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestHandler_Analyze_NotFound(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, &mockEngine{}, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/api/sessions/nonexistent/analyze", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_GetReport(t *testing.T) {
	// V10: GET /api/sessions/{id}/report returns latest report
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "report-get",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	dims, _ := json.Marshal([]DimensionResult{{Name: "test", Score: 8, Comment: "good"}})
	analysisStore.CreateReport(ctx, &store.AnalysisReport{
		SessionID:        "report-get",
		DimensionResults: string(dims),
		MarkdownContent:  "# Test Report",
		ModelUsed:        "gpt-4o",
		CreatedAt:        time.Now(),
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, nil, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/report-get/report", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp reportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Markdown != "# Test Report" {
		t.Errorf("markdown = %q, want %q", resp.Markdown, "# Test Report")
	}
}

func TestHandler_GetReport_NotFound(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "no-report",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, nil, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/no-report/report", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_ListReports(t *testing.T) {
	// V10: GET /api/sessions/{id}/reports returns history
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "list-reports",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	for range 2 {
		analysisStore.CreateReport(ctx, &store.AnalysisReport{
			SessionID:        "list-reports",
			DimensionResults: "[]",
			MarkdownContent:  "report",
			ModelUsed:        "gpt-4o",
			CreatedAt:        time.Now(),
		})
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, nil, sessionStore, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/list-reports/reports", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp listReportsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Reports) != 2 {
		t.Errorf("reports = %d, want 2", len(resp.Reports))
	}
}

func TestHandler_RoutesRegistered(t *testing.T) {
	// V12: verify route registration doesn't panic
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(nil, nil, nil, logger)
	handler := NewHandler(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
}
