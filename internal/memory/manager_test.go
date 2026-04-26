package memory

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/store"
)

type mockLLMClient struct {
	response string
	err      error
	called   bool
}

func (m *mockLLMClient) Chat(ctx context.Context, messages []llm.ChatMessage, opts *llm.ChatOptions) (*llm.ChatResponse, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{Content: m.response}, nil
}

func makeMessages(n int) []store.Message {
	msgs := make([]store.Message, n)
	for i := range msgs {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		msgs[i] = store.Message{
			SessionID:   "s1",
			Role:        role,
			Content:     "message content",
			SequenceNum: i + 1,
			CreatedAt:   time.Now(),
		}
	}
	return msgs
}

func TestBuildContext_WithinWindow(t *testing.T) {
	mock := &mockLLMClient{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := NewManager(10, mock, logger)

	msgs := makeMessages(6)
	result, err := mgr.BuildContext(context.Background(), "system prompt", msgs)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	if result.MemorySource != "window" {
		t.Errorf("MemorySource = %q, want %q", result.MemorySource, "window")
	}
	// 1 system + 6 messages
	if len(result.Messages) != 7 {
		t.Errorf("len(Messages) = %d, want 7", len(result.Messages))
	}
	if result.Messages[0].Role != llm.RoleSystem {
		t.Errorf("first message role = %q, want system", result.Messages[0].Role)
	}
	if mock.called {
		t.Error("LLM should not be called when within window")
	}
}

func TestBuildContext_OverflowTriggerSummary(t *testing.T) {
	mock := &mockLLMClient{response: "这是对话历史摘要"}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := NewManager(4, mock, logger)

	msgs := makeMessages(8) // 8 > window=4, overflow=4
	result, err := mgr.BuildContext(context.Background(), "system prompt", msgs)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	if result.MemorySource != "summary+window" {
		t.Errorf("MemorySource = %q, want %q", result.MemorySource, "summary+window")
	}
	if !mock.called {
		t.Error("LLM should be called for summary when overflow")
	}
	// 1 system (enriched) + 4 window messages
	if len(result.Messages) != 5 {
		t.Errorf("len(Messages) = %d, want 5", len(result.Messages))
	}
	// system prompt should contain summary
	if result.Messages[0].Content == "system prompt" {
		t.Error("system prompt should be enriched with summary")
	}
}

func TestBuildContext_SummaryFailureDegradation(t *testing.T) {
	mock := &mockLLMClient{err: errors.New("llm unavailable")}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := NewManager(4, mock, logger)

	msgs := makeMessages(8)
	result, err := mgr.BuildContext(context.Background(), "system prompt", msgs)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	if result.MemorySource != "window" {
		t.Errorf("MemorySource = %q, want %q (degraded)", result.MemorySource, "window")
	}
	// 1 system + 4 window
	if len(result.Messages) != 5 {
		t.Errorf("len(Messages) = %d, want 5", len(result.Messages))
	}
	// system prompt should not be enriched
	if result.Messages[0].Content != "system prompt" {
		t.Error("system prompt should remain original when summary fails")
	}
}

func TestBuildContext_EmptyMessages(t *testing.T) {
	mock := &mockLLMClient{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := NewManager(10, mock, logger)

	result, err := mgr.BuildContext(context.Background(), "system prompt", nil)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	if len(result.Messages) != 1 {
		t.Errorf("len(Messages) = %d, want 1 (system only)", len(result.Messages))
	}
}
