package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGetSession(t *testing.T) {
	db := setupTestDB(t)
	s := NewSessionStore(db)
	ctx := context.Background()

	now := time.Now()
	sess := &Session{
		ID:         "test-1",
		RoleConfig: `{"description":"面试者","scenario":"技术面试"}`,
		Goals:      `[{"name":"逻辑条理性"}]`,
		Dimensions: `[{"name":"论点清晰度"}]`,
		Status:     "active",
		RoundLimit: 10,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.CreateSession(ctx, sess); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := s.GetSession(ctx, "test-1")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.ID != "test-1" {
		t.Errorf("ID = %q, want %q", got.ID, "test-1")
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
	if got.RoundLimit != 10 {
		t.Errorf("RoundLimit = %d, want %d", got.RoundLimit, 10)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	db := setupTestDB(t)
	s := NewSessionStore(db)
	ctx := context.Background()

	got, err := s.GetSession(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got != nil {
		t.Fatal("expected nil for nonexistent session")
	}
}

func TestUpdateSessionStatus(t *testing.T) {
	db := setupTestDB(t)
	s := NewSessionStore(db)
	ctx := context.Background()

	now := time.Now()
	sess := &Session{
		ID: "test-2", RoleConfig: "{}", Goals: "[]", Dimensions: "[]",
		Status: "active", RoundLimit: 5, CreatedAt: now, UpdatedAt: now,
	}
	s.CreateSession(ctx, sess)

	if err := s.UpdateSessionStatus(ctx, "test-2", "completed"); err != nil {
		t.Fatalf("UpdateSessionStatus: %v", err)
	}

	got, _ := s.GetSession(ctx, "test-2")
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
}

func TestCreateAndListMessages(t *testing.T) {
	db := setupTestDB(t)
	s := NewSessionStore(db)
	ctx := context.Background()

	now := time.Now()
	sess := &Session{
		ID: "test-3", RoleConfig: "{}", Goals: "[]", Dimensions: "[]",
		Status: "active", RoundLimit: 0, CreatedAt: now, UpdatedAt: now,
	}
	s.CreateSession(ctx, sess)

	msgs := []*Message{
		{SessionID: "test-3", Role: "user", Content: "hello", SequenceNum: 1, CreatedAt: now},
		{SessionID: "test-3", Role: "assistant", Content: "hi there", SequenceNum: 2, CreatedAt: now},
		{SessionID: "test-3", Role: "user", Content: "how are you", SequenceNum: 3, CreatedAt: now},
	}
	for _, m := range msgs {
		if err := s.CreateMessage(ctx, m); err != nil {
			t.Fatalf("CreateMessage: %v", err)
		}
	}

	list, err := s.ListMessages(ctx, "test-3")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len(messages) = %d, want 3", len(list))
	}
	if list[0].Content != "hello" || list[2].Content != "how are you" {
		t.Errorf("messages not in order: %v", list)
	}

	count, err := s.CountMessages(ctx, "test-3")
	if err != nil {
		t.Fatalf("CountMessages: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}
