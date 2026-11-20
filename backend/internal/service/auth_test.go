package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignupValidatesEmailAndPasswordBeforeRepositoryAccess(t *testing.T) {
	auth := NewAuthService(nil, "01234567890123456789012345678901", "test", time.Hour)

	if _, err := auth.Signup(context.Background(), "not-an-email", "long-enough"); err != ErrInvalidInput {
		t.Fatalf("invalid email error = %v", err)
	}
	if _, err := auth.Signup(context.Background(), "person@example.com", "short"); err != ErrInvalidInput {
		t.Fatalf("short password error = %v", err)
	}
}

func TestNormalizeEmailTrimsAndLowercases(t *testing.T) {
	email, err := normalizeEmail(" Person@Example.com ")
	if err != nil {
		t.Fatalf("normalizeEmail() error = %v", err)
	}
	if email != "person@example.com" {
		t.Fatalf("normalized email = %q", email)
	}
}

func TestParseTokenRejectsInvalidToken(t *testing.T) {
	auth := NewAuthService(nil, "01234567890123456789012345678901", "test", time.Hour)
	if _, err := auth.ParseToken("not-a-jwt"); err == nil {
		t.Fatal("invalid token should be rejected")
	}
}

func TestIssueAndParseToken(t *testing.T) {
	auth := NewAuthService(nil, "01234567890123456789012345678901", "test", time.Hour)
	userID := "507f1f77bcf86cd799439011"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "test",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parsed, err := auth.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if parsed != userID {
		t.Fatalf("parsed subject = %q", parsed)
	}
}
