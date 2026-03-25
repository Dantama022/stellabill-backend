package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("ENV", "")
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("JWT_SECRET", "")

	cfg := Load()

	if cfg.Env != "development" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.Port != "8080" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.DBConn != "postgres://localhost/stellarbill?sslmode=disable" {
		t.Fatalf("DBConn = %q", cfg.DBConn)
	}
	if cfg.JWTSecret != "change-me-in-production" {
		t.Fatalf("JWTSecret = %q", cfg.JWTSecret)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://db/internal")
	t.Setenv("JWT_SECRET", "secret")

	cfg := Load()

	if cfg.Env != "production" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.DBConn != "postgres://db/internal" {
		t.Fatalf("DBConn = %q, want %q", cfg.DBConn, "postgres://db/internal")
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("JWTSecret = %q, want %q", cfg.JWTSecret, "secret")
	}
}
