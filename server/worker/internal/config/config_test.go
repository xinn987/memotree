package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("MEDIA_WORKER_CONCURRENCY", "")

	cfg := Load()
	if cfg.AppEnv != "local" {
		t.Fatalf("expected local env, got %q", cfg.AppEnv)
	}
	if cfg.Concurrency != 2 {
		t.Fatalf("expected default concurrency 2, got %d", cfg.Concurrency)
	}
}
