package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestAnalysisStore_CreateAndGetReport(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	sessionStore := NewSessionStore(db)
	analysisStore := NewAnalysisStore(db)

	ctx := context.Background()

	// Create a session first
	sess := &Session{
		ID:         "test-session-1",
		RoleConfig: `{"description":"test"}`,
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		RoundLimit: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Create a report
	report := &AnalysisReport{
		SessionID:        "test-session-1",
		DimensionResults: `[{"name":"test","score":8}]`,
		MarkdownContent:  "# Test Report",
		ModelUsed:        "gpt-4o",
		CreatedAt:        time.Now(),
	}
	if err := analysisStore.CreateReport(ctx, report); err != nil {
		t.Fatalf("create report: %v", err)
	}
	if report.ID == 0 {
		t.Error("expected report ID to be set after creation")
	}

	// Get latest report
	got, err := analysisStore.GetLatestReport(ctx, "test-session-1")
	if err != nil {
		t.Fatalf("get latest report: %v", err)
	}
	if got == nil {
		t.Fatal("expected report, got nil")
	}
	if got.SessionID != "test-session-1" {
		t.Errorf("session ID = %q, want %q", got.SessionID, "test-session-1")
	}
	if got.MarkdownContent != "# Test Report" {
		t.Errorf("markdown = %q, want %q", got.MarkdownContent, "# Test Report")
	}
	if got.ModelUsed != "gpt-4o" {
		t.Errorf("model_used = %q, want %q", got.ModelUsed, "gpt-4o")
	}
}

func TestAnalysisStore_ListReports(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	sessionStore := NewSessionStore(db)
	analysisStore := NewAnalysisStore(db)

	ctx := context.Background()

	sess := &Session{
		ID:         "test-session-2",
		RoleConfig: `{"description":"test"}`,
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		RoundLimit: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Create two reports
	for i := 0; i < 2; i++ {
		report := &AnalysisReport{
			SessionID:        "test-session-2",
			DimensionResults: "[]",
			MarkdownContent:  "report",
			ModelUsed:        "gpt-4o",
			CreatedAt:        time.Now(),
		}
		if err := analysisStore.CreateReport(ctx, report); err != nil {
			t.Fatalf("create report %d: %v", i, err)
		}
	}

	reports, err := analysisStore.ListReports(ctx, "test-session-2")
	if err != nil {
		t.Fatalf("list reports: %v", err)
	}
	if len(reports) != 2 {
		t.Errorf("got %d reports, want 2", len(reports))
	}
}

func TestAnalysisStore_GetLatestReport_Empty(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	analysisStore := NewAnalysisStore(db)
	ctx := context.Background()

	got, err := analysisStore.GetLatestReport(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent session")
	}
}

func TestSchemaMigration_NewColumns(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Verify markdown_content and model_used columns exist and are writable
	ctx := context.Background()
	sessionStore := NewSessionStore(db)
	analysisStore := NewAnalysisStore(db)

	sess := &Session{
		ID:         "schema-test",
		RoleConfig: "{}",
		Goals:      "[]",
		Dimensions: "[]",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := sessionStore.CreateSession(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	report := &AnalysisReport{
		SessionID:        "schema-test",
		DimensionResults: "[]",
		MarkdownContent:  "# Migration Test\nMarkdown content works!",
		ModelUsed:        "test-model",
		CreatedAt:        time.Now(),
	}
	if err := analysisStore.CreateReport(ctx, report); err != nil {
		t.Fatalf("create report with new columns: %v", err)
	}

	got, err := analysisStore.GetLatestReport(ctx, "schema-test")
	if err != nil {
		t.Fatalf("get report: %v", err)
	}
	if got.MarkdownContent != "# Migration Test\nMarkdown content works!" {
		t.Errorf("markdown_content = %q, unexpected", got.MarkdownContent)
	}
	if got.ModelUsed != "test-model" {
		t.Errorf("model_used = %q, want %q", got.ModelUsed, "test-model")
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}
