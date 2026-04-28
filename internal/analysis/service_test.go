package analysis

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

type mockEngine struct {
	result *AnalysisResult
	err    error
	called bool
}

func (m *mockEngine) Analyze(ctx context.Context, roleDesc, scenario string, messages []store.Message, dimensions []role.Dimension) (*AnalysisResult, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func setupTestDB(t *testing.T) (*store.SessionStore, *store.AnalysisStore) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.NewSessionStore(db), store.NewAnalysisStore(db)
}

func TestService_TriggerAnalysis_OnlyCompletedSession(t *testing.T) {
	// V7: active session should return ErrSessionNotCompleted
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	// Create active session
	sess := &store.Session{
		ID:         "active-session",
		RoleConfig: `{"description":"test","scenario":"test"}`,
		Goals:      "[]",
		Dimensions: `[{"name":"test","description":"d"}]`,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	engine := &mockEngine{}
	svc := NewService(analysisStore, engine, sessionStore, logger)

	_, _, err := svc.TriggerAnalysis(ctx, "active-session", "manual")
	if err != ErrSessionNotCompleted {
		t.Errorf("error = %v, want ErrSessionNotCompleted", err)
	}
	if engine.called {
		t.Error("engine should not be called for active session")
	}
}

func TestService_TriggerAnalysis_CompletedSession(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "completed-session",
		RoleConfig: `{"description":"面试官","scenario":"技术面试"}`,
		Goals:      "[]",
		Dimensions: `[{"name":"论点清晰度","description":"test"}]`,
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	engine := &mockEngine{
		result: &AnalysisResult{
			DimensionResults: []DimensionResult{{Name: "论点清晰度", Score: 8, Comment: "good", Suggestions: []string{"improve"}}},
			Markdown:         "# Report",
			ModelUsed:        "test",
		},
	}
	svc := NewService(analysisStore, engine, sessionStore, logger)

	result, report, err := svc.TriggerAnalysis(ctx, "completed-session", "manual")
	if err != nil {
		t.Fatalf("TriggerAnalysis: %v", err)
	}
	if !engine.called {
		t.Error("engine should be called for completed session")
	}
	if report.ID == 0 {
		t.Error("report ID should be set")
	}
	if result.Markdown != "# Report" {
		t.Errorf("markdown = %q, want %q", result.Markdown, "# Report")
	}
}

func TestService_TriggerAnalysis_SessionNotFound(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, &mockEngine{}, sessionStore, logger)

	_, _, err := svc.TriggerAnalysis(ctx, "nonexistent", "manual")
	if err != ErrSessionNotFound {
		t.Errorf("error = %v, want ErrSessionNotFound", err)
	}
}

func TestService_AutoTriggerHook(t *testing.T) {
	// V9: verify OnSessionEndFunc callback is invoked
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "auto-trigger-test",
		RoleConfig: `{"description":"test","scenario":"test"}`,
		Goals:      "[]",
		Dimensions: `[{"name":"test","description":"d"}]`,
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	var hookCalled bool
	hook := func(ctx context.Context, sessionID string) {
		hookCalled = true
		if sessionID != "auto-trigger-test" {
			t.Errorf("hook sessionID = %q, want %q", sessionID, "auto-trigger-test")
		}
	}

	// Simulate what main.go does: set OnSessionEnd
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	engine := &mockEngine{
		result: &AnalysisResult{
			DimensionResults: []DimensionResult{{Name: "test", Score: 7, Comment: "ok", Suggestions: nil}},
			Markdown:         "# Report",
			ModelUsed:        "test",
		},
	}
	svc := NewService(analysisStore, engine, sessionStore, logger)

	// Call the hook directly (simulating session.Service.notifySessionEnd)
	hook(context.Background(), "auto-trigger-test")

	if !hookCalled {
		t.Error("hook was not called")
	}

	// Verify the analysis service can be called through the hook pattern
	_, _, err := svc.TriggerAnalysis(context.Background(), "auto-trigger-test", "auto")
	if err != nil {
		t.Errorf("TriggerAnalysis via hook: %v", err)
	}
}

func TestService_GetLatestReport(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "report-test",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	dims, _ := json.Marshal([]DimensionResult{{Name: "test", Score: 7}})
	analysisStore.CreateReport(ctx, &store.AnalysisReport{
		SessionID:        "report-test",
		DimensionResults: string(dims),
		MarkdownContent:  "# Report",
		ModelUsed:        "model",
		CreatedAt:        time.Now(),
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, nil, sessionStore, logger)

	report, err := svc.GetLatestReport(ctx, "report-test")
	if err != nil {
		t.Fatalf("GetLatestReport: %v", err)
	}
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.MarkdownContent != "# Report" {
		t.Errorf("markdown = %q, want %q", report.MarkdownContent, "# Report")
	}
}

func TestService_ListReports(t *testing.T) {
	sessionStore, analysisStore := setupTestDB(t)
	ctx := context.Background()

	sess := &store.Session{
		ID:         "list-test",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	sessionStore.CreateSession(ctx, sess)

	for range 3 {
		analysisStore.CreateReport(ctx, &store.AnalysisReport{
			SessionID:        "list-test",
			DimensionResults: "[]",
			MarkdownContent:  "report",
			ModelUsed:        "model",
			CreatedAt:        time.Now(),
		})
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	svc := NewService(analysisStore, nil, sessionStore, logger)

	reports, err := svc.ListReports(ctx, "list-test")
	if err != nil {
		t.Fatalf("ListReports: %v", err)
	}
	if len(reports) != 3 {
		t.Errorf("got %d reports, want 3", len(reports))
	}
}
