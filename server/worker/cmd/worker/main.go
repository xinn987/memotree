// Worker 进程入口：装配数据库、对象存储和媒体处理调度器。
package main

import (
	"context"
	"database/sql"
	"log"

	"memotree/server/internal/logging"
	"memotree/server/internal/storage"
	"memotree/server/worker/internal/config"
	"memotree/server/worker/internal/media"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cleanupLog, err := logging.ConfigureFromEnv("worker")
	if err != nil {
		log.Fatalf("configure logging: %v", err)
	}
	defer cleanupLog()

	cfg := config.Load()
	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN is required for media worker")
	}
	if cfg.StorageAccessKeyID == "" || cfg.StorageSecretKey == "" {
		log.Fatal("storage credentials are required for media worker")
	}

	ctx := context.Background()
	videoTools, err := media.CheckVideoTools(ctx, cfg.FFmpegPath, cfg.FFprobePath)
	if err != nil {
		log.Fatalf("video processing dependencies are unavailable: %v", err)
	}
	log.Printf("video processing dependencies ready ffmpeg=%q ffprobe=%q", videoTools.FFmpeg, videoTools.FFprobe)

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping mysql: %v", err)
	}

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

	repository := media.NewMySQLRepository(db, 30*cfg.PollInterval)
	processor := media.Processor{
		Repository:      repository,
		ObjectStore:     media.S3ObjectStore{Service: s3Storage},
		OriginalsBucket: cfg.OriginalsBucket,
		PreviewsBucket:  cfg.PreviewsBucket,
		VideoTranscoder: media.FFmpegTranscoder{
			Path:        cfg.FFmpegPath,
			FFprobePath: cfg.FFprobePath,
		},
	}
	runner := media.Runner{
		Repository:   repository,
		Processor:    processor,
		BatchSize:    cfg.Concurrency,
		PollInterval: cfg.PollInterval,
		Logger:       log.Default(),
	}

	log.Printf("memotree worker started env=%s concurrency=%d poll=%s", cfg.AppEnv, cfg.Concurrency, cfg.PollInterval)
	if err := runner.Run(ctx, func(err error) {
		log.Printf("media worker tick failed: %v", err)
	}); err != nil {
		log.Fatalf("media worker stopped: %v", err)
	}
}
