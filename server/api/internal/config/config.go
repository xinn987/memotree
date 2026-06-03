package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv              string
	APIAddr             string
	MySQLDSN            string
	StorageProvider     string
	StorageEndpoint     string
	StorageRegion       string
	OriginalsBucket     string
	PreviewsBucket      string
	SignedURLTTL         time.Duration
	UploadMaxFileBytes  int64
	UploadMaxBatchCount int
}

func Load() Config {
	return Config{
		AppEnv:              getEnv("APP_ENV", "local"),
		APIAddr:             getEnv("API_ADDR", ":8080"),
		MySQLDSN:            getEnv("MYSQL_DSN", ""),
		StorageProvider:     getEnv("STORAGE_PROVIDER", "r2"),
		StorageEndpoint:     getEnv("STORAGE_ENDPOINT", ""),
		StorageRegion:       getEnv("STORAGE_REGION", "auto"),
		OriginalsBucket:     getEnv("STORAGE_BUCKET_ORIGINALS", "memotree-originals"),
		PreviewsBucket:      getEnv("STORAGE_BUCKET_PREVIEWS", "memotree-previews"),
		SignedURLTTL:         time.Duration(getEnvInt("SIGNED_URL_TTL_SECONDS", 900)) * time.Second,
		UploadMaxFileBytes:  getEnvInt64("UPLOAD_MAX_FILE_BYTES", 1073741824),
		UploadMaxBatchCount: getEnvInt("UPLOAD_MAX_BATCH_COUNT", 50),
	}
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

func getEnvInt64(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(getEnv(key, ""), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
