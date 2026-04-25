package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lq5657/talkent/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

func newTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newTestClient(ts *httptest.Server, logger *slog.Logger) *OpenAIClient {
	cfg := &config.LLMConfig{
		Provider: "test",
		BaseURL:  ts.URL,
		APIKey:   "test-api-key",
		Model:    "test-model",
		Timeout:  10 * time.Second,
	}

	ocfg := openai.DefaultConfig(cfg.APIKey)
	ocfg.BaseURL = ts.URL

	return &OpenAIClient{
		client: openai.NewClientWithConfig(ocfg),
		cfg:    cfg,
		logger: logger,
	}
}

func chatCompletionResponse(model, content string, prompt, completion, total int) string {
	return fmt.Sprintf(`{
		"id": "chatcmpl-test",
		"object": "chat.completion",
		"model": "%s",
		"choices": [{"index":0,"message":{"role":"assistant","content":"%s"},"finish_reason":"stop"}],
		"usage": {"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d}
	}`, model, content, prompt, completion, total)
}

func TestChatCompletion_Success(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request: %v", err)
		}

		msgs, _ := req["messages"].([]any)
		if len(msgs) != 2 {
			t.Errorf("expected 2 messages, got %d", len(msgs))
		}
		if req["model"] != "test-model" {
			t.Errorf("expected model test-model, got %v", req["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "Hello from AI", 10, 20, 30)))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	resp, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleSystem, Content: "You are a helpful assistant."},
		{Role: RoleUser, Content: "Hello"},
	}, &ChatOptions{Temperature: 0.7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello from AI" {
		t.Errorf("expected 'Hello from AI', got %s", resp.Content)
	}
	if resp.Model != "test-model" {
		t.Errorf("expected model test-model, got %s", resp.Model)
	}
	if resp.TotalTokens != 30 {
		t.Errorf("expected 30 total tokens, got %d", resp.TotalTokens)
	}
}

func TestChatCompletion_CustomBaseURL(t *testing.T) {
	receivedURL := ""
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		receivedURL = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "ok", 1, 1, 2)))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	_, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedURL != "/chat/completions" {
		t.Errorf("request not sent to test server, got path: %s", receivedURL)
	}
}

func TestChatCompletion_RetryOn429(t *testing.T) {
	callCount := 0
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"message":"rate limit","type":"rate_limit_error","code":429}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "retried", 1, 1, 2)))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	resp, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp.Content != "retried" {
		t.Errorf("expected 'retried', got %s", resp.Content)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls (2 failures + 1 success), got %d", callCount)
	}
}

func TestChatCompletion_RetryOn500(t *testing.T) {
	callCount := 0
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"internal error","type":"server_error","code":500}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "ok", 1, 1, 2)))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	resp, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("expected 'ok', got %s", resp.Content)
	}
}

func TestChatCompletion_NoRetryOn401(t *testing.T) {
	callCount := 0
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid api key","type":"authentication_error","code":401}}`))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	_, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry on 401), got %d", callCount)
	}
}

func TestChatCompletion_RetryExhausted(t *testing.T) {
	callCount := 0
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"message":"overloaded","type":"server_error","code":503}}`))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	_, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err == nil {
		t.Fatal("expected error after retries exhausted, got nil")
	}
	// 1 initial + maxRetries(3) = 4
	if callCount != 4 {
		t.Errorf("expected 4 calls (1 initial + 3 retries), got %d", callCount)
	}
}

func TestChatCompletion_ContextCancellation(t *testing.T) {
	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "slow", 1, 1, 2)))
	})
	defer ts.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client := newTestClient(ts, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.Chat(ctx, []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err == nil {
		t.Fatal("expected error from context cancellation, got nil")
	}
}

func TestChatCompletion_APIKeyNotInLogs(t *testing.T) {
	var logOutput strings.Builder
	logger := slog.New(slog.NewTextHandler(&logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ts := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(chatCompletionResponse("test-model", "ok", 1, 1, 2)))
	})
	defer ts.Close()

	client := newTestClient(ts, logger)

	_, err := client.Chat(context.Background(), []ChatMessage{
		{Role: RoleUser, Content: "test"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs := logOutput.String()
	if strings.Contains(logs, "test-api-key") {
		t.Errorf("API key found in logs:\n%s", logs)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"429 rate limit", 429, true},
		{"500 internal error", 500, true},
		{"502 bad gateway", 502, true},
		{"503 service unavailable", 503, true},
		{"401 unauthorized", 401, false},
		{"403 forbidden", 403, false},
		{"400 bad request", 400, false},
		{"404 not found", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &openai.APIError{HTTPStatusCode: tt.statusCode}
			if got := isRetryableError(err); got != tt.want {
				t.Errorf("isRetryableError(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}

	t.Run("non-API error", func(t *testing.T) {
		if isRetryableError(errors.New("some error")) {
			t.Error("isRetryableError should return false for non-API errors")
		}
	})
}

func TestNewClient_NilConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := NewClient(nil, logger)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
}
