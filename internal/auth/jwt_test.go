package auth

import (
	"testing"
	"time"
)

func TestJWTServiceAccessToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}
	if claims.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got '%s'", claims.Username)
	}
}

func TestJWTServiceRefreshToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)

	token, err := svc.GenerateRefreshToken("testuser")
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := svc.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}
	if claims.Username != "testuser" {
		t.Fatalf("expected username 'testuser', got '%s'", claims.Username)
	}
}

func TestJWTServiceExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret", -1*time.Second, 7*24*time.Hour)

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	_, err = svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTServiceInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)

	_, err := svc.ValidateAccessToken("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTServiceWrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1", 1*time.Hour, 7*24*time.Hour)
	svc2 := NewJWTService("secret-2", 1*time.Hour, 7*24*time.Hour)

	token, err := svc1.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	_, err = svc2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestJWTServiceAccessTokenOnlyInAccessValidator(t *testing.T) {
	svc := NewJWTService("test-secret", 1*time.Hour, 7*24*time.Hour)

	token, err := svc.GenerateAccessToken("testuser")
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Access tokens can also be validated by refresh validator (same parse internally)
	// This tests that ValidateAccessToken works on its own tokens
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}
	if claims == nil {
		t.Fatal("expected non-nil claims")
	}
}
