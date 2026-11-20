package service

import (
	"context"
	"testing"
	"time"
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
