// init-storage 是本地开发运维命令，用于幂等创建对象存储 bucket。
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"memotree/server/devtools/internal/buckets"
	"memotree/server/internal/storage"
)

func main() {
	cfg := loadConfig()
	ctx := context.Background()
	s3Storage, err := storage.NewS3Service(storage.S3Config{
		Endpoint:        cfg.StorageEndpoint,
		Region:          cfg.StorageRegion,
		AccessKeyID:     cfg.StorageAccessKeyID,
		SecretAccessKey: cfg.StorageSecretKey,
		UsePathStyle:    cfg.StorageUsePathStyle,
	})
	if err != nil {
		log.Fatalf("configure storage: %v", err)
	}
	if err := buckets.EnsureAll(ctx, s3Storage, []string{cfg.OriginalsBucket, cfg.PreviewsBucket}); err != nil {
		log.Fatalf("ensure storage buckets: %v", err)
	}
	log.Printf("storage buckets are ready: %s, %s", cfg.OriginalsBucket, cfg.PreviewsBucket)
}

type config struct {
	StorageEndpoint     string
	StorageRegion       string
	StorageAccessKeyID  string
	StorageSecretKey    string
	StorageUsePathStyle bool
	OriginalsBucket     string
	PreviewsBucket      string
}

func loadConfig() config {
	return config{
		StorageEndpoint:     getEnv("STORAGE_ENDPOINT", ""),
		StorageRegion:       getEnv("STORAGE_REGION", "auto"),
		StorageAccessKeyID:  getEnv("STORAGE_ACCESS_KEY_ID", ""),
		StorageSecretKey:    getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
		StorageUsePathStyle: getEnvBool("STORAGE_USE_PATH_STYLE", false),
		OriginalsBucket:     getEnv("STORAGE_BUCKET_ORIGINALS", "memotree-originals"),
		PreviewsBucket:      getEnv("STORAGE_BUCKET_PREVIEWS", "memotree-previews"),
	}
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.ToLower(getEnv(key, ""))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
