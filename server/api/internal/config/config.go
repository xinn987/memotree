// Package config 读取 API 进程运行配置。
//
// 当前只读取环境变量；.env.example 作为配置样例，不由代码自动加载。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 是 API 运行时配置快照。
// 字段保持扁平，方便 main 按模块把配置传给 HTTP、store 和 storage。
type Config struct {
	AppEnv              string
	APIAddr             string
	MySQLDSN            string
	SessionCookieName   string
	SessionTTL          time.Duration
	StorageProvider     string
	StorageEndpoint     string
	StorageRegion       string
	StorageAccessKeyID  string
	StorageSecretKey    string
	StorageUsePathStyle bool
	OriginalsBucket     string
	PreviewsBucket      string
	SignedURLTTL        time.Duration
	UploadMaxFileBytes  int64
	UploadMaxBatchCount int
}

// Load 从环境变量读取配置，并为本地开发提供保守默认值。
func Load() Config {
	return Config{
		AppEnv:              getEnv("APP_ENV", "local"),
		APIAddr:             getEnv("API_ADDR", ":8080"),
		MySQLDSN:            getEnv("MYSQL_DSN", ""),
		SessionCookieName:   getEnv("SESSION_COOKIE_NAME", "memotree_session"),
		SessionTTL:          time.Duration(getEnvInt("SESSION_TTL_HOURS", 24*30)) * time.Hour,
		StorageProvider:     getEnv("STORAGE_PROVIDER", "r2"),
		StorageEndpoint:     getEnv("STORAGE_ENDPOINT", ""),
		StorageRegion:       getEnv("STORAGE_REGION", "auto"),
		StorageAccessKeyID:  getEnv("STORAGE_ACCESS_KEY_ID", ""),
		StorageSecretKey:    getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
		StorageUsePathStyle: getEnvBool("STORAGE_USE_PATH_STYLE", false),
		OriginalsBucket:     getEnv("STORAGE_BUCKET_ORIGINALS", "memotree-originals"),
		PreviewsBucket:      getEnv("STORAGE_BUCKET_PREVIEWS", "memotree-previews"),
		SignedURLTTL:        time.Duration(getEnvInt("SIGNED_URL_TTL_SECONDS", 900)) * time.Second,
		UploadMaxFileBytes:  getEnvInt64("UPLOAD_MAX_FILE_BYTES", 1073741824),
		UploadMaxBatchCount: getEnvInt("UPLOAD_MAX_BATCH_COUNT", 50),
	}
}

// ValidateRuntimeDependencies 校验类生产环境的关键依赖，避免误用内存库或禁用对象存储。
func (c Config) ValidateRuntimeDependencies() error {
	if !isProductionLike(c.AppEnv) {
		return nil
	}
	if strings.TrimSpace(c.MySQLDSN) == "" {
		return fmt.Errorf("MYSQL_DSN is required when APP_ENV=%s", c.AppEnv)
	}
	if strings.TrimSpace(c.StorageAccessKeyID) == "" || strings.TrimSpace(c.StorageSecretKey) == "" {
		return fmt.Errorf("storage credentials are required when APP_ENV=%s", c.AppEnv)
	}
	if strings.TrimSpace(c.StorageEndpoint) == "" {
		return fmt.Errorf("STORAGE_ENDPOINT is required when APP_ENV=%s", c.AppEnv)
	}
	if strings.TrimSpace(c.OriginalsBucket) == "" || strings.TrimSpace(c.PreviewsBucket) == "" {
		return fmt.Errorf("storage bucket names are required when APP_ENV=%s", c.AppEnv)
	}
	return nil
}

func isProductionLike(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "", "local", "test":
		return false
	default:
		return true
	}
}

// getEnv 读取字符串环境变量，空字符串视为未配置。
func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt 读取整数环境变量，解析失败时回退默认值。
func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

// getEnvInt64 读取 int64 环境变量，主要用于文件大小限制。
func getEnvInt64(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(getEnv(key, ""), 10, 64)
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
