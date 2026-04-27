package session

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/memory"
	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

type mockClient struct {
	response string
	err      error
}

func (m *mockClient) Chat(ctx context.Context, messages []llm.ChatMessage, opts *llm.ChatOptions) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{Content: m.response}, nil
}

func setupTestStore(t *testing.T) *store.SessionStore {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.NewSessionStore(db)
}

var testLogger = slog.New(slog.NewTextHandler(os.Stderr, nil))

func TestCreateSession(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "你好"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	req := CreateSessionRequest{
		RoleDescription: "面试者",
		Scenario:        "技术面试",
		RoleType:        role.RoleTypeStructuredExpression,
		Goals:           []role.TrainingGoal{{Name: "逻辑条理性", Description: "观点是否有逻辑链条"}},
		Dimensions:      []role.Dimension{{Name: "论点清晰度", Description: "核心论点是否明确"}},
		RoundLimit:      10,
	}

	sess, err := svc.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if sess.Status != "active" {
		t.Errorf("Status = %q, want %q", sess.Status, "active")
	}
	if sess.RoundLimit != 10 {
		t.Errorf("RoundLimit = %d, want 10", sess.RoundLimit)
	}

	// Verify role_config stored correctly
	var rc roleConfigJSON
	json.Unmarshal([]byte(sess.RoleConfig), &rc)
	if rc.Description != "面试者" {
		t.Errorf("RoleConfig.Description = %q, want %q", rc.Description, "面试者")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	prompt := BuildSystemPrompt(
		"面试者",
		"技术面试",
		[]role.TrainingGoal{{Name: "逻辑条理性", Description: "观点是否有逻辑链条"}},
		[]role.Dimension{{Name: "论点清晰度", Description: "核心论点是否明确"}},
	)

	if prompt == "" {
		t.Fatal("expected non-empty prompt")
	}
	// Must contain key parts
	checks := []string{"面试者", "技术面试", "逻辑条理性", "论点清晰度"}
	for _, c := range checks {
		if !strings.Contains(prompt, c) {
			t.Errorf("prompt missing %q", c)
		}
	}
}

func TestChat_EndToEnd(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "你好，请坐"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	sess, _ := svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		Scenario:        "技术面试",
		RoundLimit:      10,
	})

	result, err := svc.Chat(context.Background(), sess.ID, "你好")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if result.Reply != "你好，请坐" {
		t.Errorf("Reply = %q, want %q", result.Reply, "你好，请坐")
	}
	if result.CurrentRound != 1 {
		t.Errorf("CurrentRound = %d, want 1", result.CurrentRound)
	}
	if result.IsLast {
		t.Error("IsLast should be false on round 1 of 10")
	}
}

func TestChat_RoundLimitAutoEnd(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "回复"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	sess, _ := svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      2,
	})

	// Round 1
	r1, _ := svc.Chat(context.Background(), sess.ID, "msg1")
	if r1.IsLast {
		t.Error("round 1 should not be last")
	}

	// Round 2 = limit reached
	r2, err := svc.Chat(context.Background(), sess.ID, "msg2")
	if err != nil {
		t.Fatalf("Chat round 2: %v", err)
	}
	if !r2.IsLast {
		t.Error("round 2 should be last (round_limit=2)")
	}

	// Verify session completed
	got, _ := s.GetSession(context.Background(), sess.ID)
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}
}

func TestChat_UnlimitedRounds(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "回复"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	sess, _ := svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      0, // unlimited
	})

	// Multiple rounds should never trigger isLast
	for i := 0; i < 5; i++ {
		r, err := svc.Chat(context.Background(), sess.ID, "msg")
		if err != nil {
			t.Fatalf("Chat round %d: %v", i+1, err)
		}
		if r.IsLast {
			t.Errorf("round %d: IsLast should be false when round_limit=0", i+1)
		}
	}

	// Session should still be active
	got, _ := s.GetSession(context.Background(), sess.ID)
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q (unlimited should stay active)", got.Status, "active")
	}
}

func TestChat_CompletedSessionRejected(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "回复"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	sess, _ := svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      1,
	})

	// Exhaust round limit
	svc.Chat(context.Background(), sess.ID, "msg1")

	// Try to chat again
	_, err := svc.Chat(context.Background(), sess.ID, "msg2")
	if err != ErrSessionCompleted {
		t.Errorf("error = %v, want ErrSessionCompleted", err)
	}
}

func TestEndSession(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{response: "回复"}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	sess, _ := svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      10,
	})

	ended, err := svc.EndSession(context.Background(), sess.ID)
	if err != nil {
		t.Fatalf("EndSession: %v", err)
	}
	if ended.Status != "completed" {
		t.Errorf("Status = %q, want %q", ended.Status, "completed")
	}

	// Double end should fail
	_, err = svc.EndSession(context.Background(), sess.ID)
	if err != ErrSessionCompleted {
		t.Errorf("double end error = %v, want ErrSessionCompleted", err)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	s := setupTestStore(t)
	mock := &mockClient{}
	mgr := memory.NewManager(10, mock, testLogger)
	svc := NewService(s, mgr, mock, testLogger)

	_, err := svc.GetSession(context.Background(), "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("error = %v, want ErrSessionNotFound", err)
	}
}

