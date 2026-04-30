package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRequestIDMiddleware(t *testing.T) {
	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestIDFromContext(r.Context())
		if id == "" {
			t.Error("expected non-empty request id in context")
		}
		respID := w.Header().Get("X-Request-ID")
		if respID == "" {
			t.Error("expected X-Request-ID header")
		}
		if id != respID {
			t.Errorf("context id %q != header id %q", id, respID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID in response header")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	logger := newTestLogger()
	handler := RecoveryMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body != `{"error":"internal server error"}` {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	logger := newTestLogger()
	handler := RecoveryMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_WithinTimeout(t *testing.T) {
	handler := TimeoutMiddleware(1 * time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestTimeoutMiddleware_Exceeded(t *testing.T) {
	handler := TimeoutMiddleware(10 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", rec.Code)
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	id := RequestIDFromContext(req.Context())
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestRecoveryInsideTimeout_PanicCaught(t *testing.T) {
	// Regression guard for C1: Recovery must be inside Timeout's goroutine
	// so that handler panics are caught instead of crashing the process.
	logger := newTestLogger()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("handler panic inside timeout goroutine")
	})
	handler = RecoveryMiddleware(logger)(handler) // innermost
	handler = TimeoutMiddleware(1 * time.Second)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if rec.Body.String() != `{"error":"internal server error"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}
