package session

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/memory"
	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

func setupHandler(t *testing.T) (*Handler, *store.SessionStore) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock := &mockClient{response: "AI 回复"}
	s := store.NewSessionStore(db)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := memory.NewManager(10, mock, logger)
	svc := NewService(s, mgr, mock, logger)
	h := NewHandler(svc, logger)
	return h, s
}

func TestHandleCreate(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(createSessionRequest{
		RoleDescription: "面试者",
		Scenario:        "技术面试",
		RoleType:        role.RoleTypeStructuredExpression,
		Goals:           []role.TrainingGoal{{Name: "逻辑条理性", Description: "test"}},
		RoundLimit:      10,
	})

	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp createSessionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SessionID == "" {
		t.Error("expected non-empty session_id")
	}
	if resp.Status != "active" {
		t.Errorf("status = %q, want %q", resp.Status, "active")
	}
}

func TestHandleCreate_MissingFields(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Missing role_description
	body, _ := json.Marshal(map[string]string{"scenario": "test"})
	req := httptest.NewRequest("POST", "/api/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleChat(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Create session first
	sess, _ := h.svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      10,
	})

	body, _ := json.Marshal(chatRequest{Content: "你好"})
	req := httptest.NewRequest("POST", "/api/sessions/"+sess.ID+"/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", sess.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp chatResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Reply != "AI 回复" {
		t.Errorf("Reply = %q, want %q", resp.Reply, "AI 回复")
	}
}

func TestHandleChat_SessionNotFound(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body, _ := json.Marshal(chatRequest{Content: "hello"})
	req := httptest.NewRequest("POST", "/api/sessions/nonexistent/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleEnd(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	sess, _ := h.svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      10,
	})

	req := httptest.NewRequest("POST", "/api/sessions/"+sess.ID+"/end", nil)
	req.SetPathValue("id", sess.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp endSessionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "completed" {
		t.Errorf("Status = %q, want %q", resp.Status, "completed")
	}
}

func TestHandleGet(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	sess, _ := h.svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		Scenario:        "技术面试",
		RoundLimit:      10,
	})

	req := httptest.NewRequest("GET", "/api/sessions/"+sess.ID, nil)
	req.SetPathValue("id", sess.ID)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp getSessionResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SessionID != sess.ID {
		t.Errorf("SessionID = %q, want %q", resp.SessionID, sess.ID)
	}
	if resp.RoleDescription != "面试者" {
		t.Errorf("RoleDescription = %q, want %q", resp.RoleDescription, "面试者")
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	h, _ := setupHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// Ensure mockClient satisfies llm.Client in this package too
var _ llm.Client = (*mockClient)(nil)

func setupStreamHandler(t *testing.T) (*Handler, *store.SessionStore) {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mock := &streamMockClient{tokens: []string{"Hello", " from", " SSE"}}
	s := store.NewSessionStore(db)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	mgr := memory.NewManager(10, mock, logger)
	svc := NewService(s, mgr, mock, logger)
	h := NewHandler(svc, logger)
	return h, s
}

func TestHandleChatStream_Success(t *testing.T) {
	h, _ := setupStreamHandler(t)

	sess, _ := h.svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
		RoundLimit:      5,
	})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/"+sess.ID+"/chat/stream?content=hello", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	body := w.Body.String()
	if !containsStr(body, `data: {"token":"Hello"}`) {
		t.Error("expected token 'Hello' in SSE stream")
	}
	if !containsStr(body, `"done":true`) {
		t.Error("expected done event in SSE stream")
	}
	if !containsStr(body, `"reply":"Hello from SSE"`) {
		t.Error("expected full reply in done event")
	}
}

func TestHandleChatStream_MissingContent(t *testing.T) {
	h, _ := setupStreamHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/some-id/chat/stream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleChatStream_SessionNotFound(t *testing.T) {
	h, _ := setupStreamHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/nonexistent/chat/stream?content=hello", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandleChatStream_SessionCompleted(t *testing.T) {
	h, s := setupStreamHandler(t)

	sess, _ := h.svc.CreateSession(context.Background(), CreateSessionRequest{
		RoleDescription: "面试者",
	})
	s.UpdateSessionStatus(context.Background(), sess.ID, "completed")

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/api/sessions/"+sess.ID+"/chat/stream?content=hello", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
