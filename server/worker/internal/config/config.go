package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv              string
	MySQLDSN            string
	Concurrency         int
	PollInterval        time.Duration
	StorageEndpoint     string
	StorageRegion       string
	StorageAccessKeyID  string
	StorageSecretKey    string
	StorageUsePathStyle bool
	OriginalsBucket     string
	PreviewsBucket      string
	FFmpegPath          string
	FFprobePath         string
}

func Load() Config {
	return Config{
		AppEnv:              getEnv("APP_ENV", "local"),
		MySQLDSN:            getEnv("MYSQL_DSN", ""),
		Concurrency:         getEnvInt("MEDIA_WORKER_CONCURRENCY", 2),
		PollInterval:        time.Duration(getEnvInt("MEDIA_WORKER_POLL_INTERVAL_SECONDS", 5)) * time.Second,
		StorageEndpoint:     getEnv("STORAGE_ENDPOINT", ""),
		StorageRegion:       getEnv("STORAGE_REGION", "auto"),
		StorageAccessKeyID:  getEnv("STORAGE_ACCESS_KEY_ID", ""),
		StorageSecretKey:    getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
		StorageUsePathStyle: getEnvBool("STORAGE_USE_PATH_STYLE", false),
		OriginalsBucket:     getEnv("STORAGE_BUCKET_ORIGINALS", "memotree-originals"),
		PreviewsBucket:      getEnv("STORAGE_BUCKET_PREVIEWS", "memotree-previews"),
		FFmpegPath:          getEnv("FFMPEG_PATH", "ffmpeg"),
		FFprobePath:         getEnv("FFPROBE_PATH", "ffprobe"),
	}
}

// ValidateRuntimeDependencies 校验 Worker 必需依赖，启动早失败比静默空跑更容易排障。
func (c Config) ValidateRuntimeDependencies() error {
	if strings.TrimSpace(c.MySQLDSN) == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}
	if strings.TrimSpace(c.StorageAccessKeyID) == "" || strings.TrimSpace(c.StorageSecretKey) == "" {
		return fmt.Errorf("storage credentials are required")
	}
	if strings.TrimSpace(c.StorageEndpoint) == "" {
		return fmt.Errorf("STORAGE_ENDPOINT is required")
	}
	if strings.TrimSpace(c.OriginalsBucket) == "" || strings.TrimSpace(c.PreviewsBucket) == "" {
		return fmt.Errorf("storage bucket names are required")
	}
	return nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.ToLower(getEnv(key, ""))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
