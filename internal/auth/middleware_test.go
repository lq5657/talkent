package auth

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestAuthMiddlewareNoToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := UsernameFromContext(r)
		if username != "testuser" {
			t.Fatalf("expected 'testuser' in context, got '%s'", username)
		}
		w.WriteHeader(http.StatusOK)
	}))

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret", -1*time.Second, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", rec.Code)
	}
}

func TestAuthMiddlewareOptionsRequest(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/sessions", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", rec.Code)
	}
}

func TestAuthMiddlewareHealthEndpoint(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", rec.Code)
	}
}

func TestAuthMiddlewareQueryParamToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)
	mw := AuthMiddleware(svc, testLogger())

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions/chat/stream?token="+token, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for query param token, got %d", rec.Code)
	}
}
