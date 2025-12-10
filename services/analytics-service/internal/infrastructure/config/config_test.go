package config

import "testing"

func TestLoadValidation(t *testing.T) {
	t.Setenv("ANALYTICS_SERVICE_DB_HOST", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestLoadSuccess(t *testing.T) {
	t.Setenv("ANALYTICS_SERVICE_DB_HOST", "localhost")
	t.Setenv("ANALYTICS_SERVICE_DB_USER", "user")
	t.Setenv("ANALYTICS_SERVICE_DB_PASSWORD", "pass")
	t.Setenv("ANALYTICS_SERVICE_DB_NAME", "db")
	t.Setenv("ANALYTICS_SERVICE_SERVICE_NAME", "custom") // typo to ensure fallback not used
	t.Setenv("ANALYTICS_SERVICE_GRPC_ADDR", ":9000")
	t.Setenv("ANALYTICS_SERVICE_HTTP_ADDR", ":8000")
	t.Setenv("ANALYTICS_SERVICE_DB_MAX_CONNS", "15")
	t.Setenv("ANALYTICS_SERVICE_DB_MIN_CONNS", "5")
	t.Setenv("ANALYTICS_SERVICE_DB_CONN_MAX_LIFETIME", "1h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.PostgresURL() != "postgres://user:pass@localhost:5432/db?sslmode=disable" {
		t.Fatalf("unexpected postgres url: %s", cfg.PostgresURL())
	}
	if cfg.Postgres.MaxConns != 15 || cfg.Postgres.MinConns != 5 {
		t.Fatalf("unexpected pool settings: %+v", cfg.Postgres)
	}
	if cfg.Postgres.MaxLifetime.String() != "1h0m0s" {
		t.Fatalf("unexpected lifetime: %s", cfg.Postgres.MaxLifetime)
	}
}

func TestParseHelpers(t *testing.T) {
	if parseDuration("bad") != 0 {
		t.Fatalf("expected bad duration to return zero")
	}
	if parseInt("oops") != 0 {
		t.Fatalf("expected bad int to return zero")
	}
	if valueOrDefault("MISSING_KEY", "fallback") != "fallback" {
		t.Fatalf("expected fallback value")
	}
}
