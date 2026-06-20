package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("MEDIA_WORKER_CONCURRENCY", "")
	t.Setenv("MEDIA_WORKER_POLL_INTERVAL_SECONDS", "")
	t.Setenv("FFMPEG_PATH", "")
	t.Setenv("FFPROBE_PATH", "")

	cfg := Load()
	if cfg.AppEnv != "local" {
		t.Fatalf("expected local env, got %q", cfg.AppEnv)
	}
	if cfg.Concurrency != 2 {
		t.Fatalf("expected default concurrency 2, got %d", cfg.Concurrency)
	}
	if cfg.PollInterval != 5*time.Second {
		t.Fatalf("expected default poll interval 5s, got %s", cfg.PollInterval)
	}
	if cfg.OriginalsBucket != "memotree-originals" || cfg.PreviewsBucket != "memotree-previews" {
		t.Fatalf("unexpected default buckets: %#v", cfg)
	}
	if cfg.FFmpegPath != "ffmpeg" || cfg.FFprobePath != "ffprobe" {
		t.Fatalf("unexpected default ffmpeg config: %#v", cfg)
	}
}

func TestLoadStorageAndDatabaseConfig(t *testing.T) {
	t.Setenv("MYSQL_DSN", "memotree:test@tcp(127.0.0.1:3307)/memotree?parseTime=true")
	t.Setenv("STORAGE_ENDPOINT", "http://127.0.0.1:9000")
	t.Setenv("STORAGE_REGION", "us-east-1")
	t.Setenv("STORAGE_ACCESS_KEY_ID", "memotree")
	t.Setenv("STORAGE_SECRET_ACCESS_KEY", "secret")
	t.Setenv("STORAGE_USE_PATH_STYLE", "true")
	t.Setenv("MEDIA_WORKER_POLL_INTERVAL_SECONDS", "3")
	t.Setenv("FFMPEG_PATH", "C:\\tools\\ffmpeg.exe")
	t.Setenv("FFPROBE_PATH", "C:\\tools\\ffprobe.exe")

	cfg := Load()
	if cfg.MySQLDSN == "" || cfg.StorageEndpoint == "" || cfg.StorageAccessKeyID == "" || cfg.StorageSecretKey == "" {
		t.Fatalf("expected database and storage config to be loaded, got %#v", cfg)
	}
	if !cfg.StorageUsePathStyle {
		t.Fatalf("expected path-style storage to be enabled")
	}
	if cfg.PollInterval != 3*time.Second {
		t.Fatalf("expected poll interval 3s, got %s", cfg.PollInterval)
	}
	if cfg.FFmpegPath != "C:\\tools\\ffmpeg.exe" || cfg.FFprobePath != "C:\\tools\\ffprobe.exe" {
		t.Fatalf("expected ffmpeg config to be loaded, got %#v", cfg)
	}
}
