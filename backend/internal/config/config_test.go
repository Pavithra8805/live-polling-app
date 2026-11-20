package config

import "testing"

func TestLoadRejectsShortJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "too-short")
	if _, err := Load(); err == nil {
		t.Fatal("expected short JWT secret to be rejected")
	}
}

func TestLoadUsesConfiguredValues(t *testing.T) {
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("JWT_EXPIRY", "2h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.JWTExpiry.String() != "2h0m0s" {
		t.Fatalf("JWTExpiry = %s", cfg.JWTExpiry)
	}
}
