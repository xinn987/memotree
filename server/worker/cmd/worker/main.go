package main

import (
	"log"

	"memotree/server/worker/internal/config"
)

func main() {
	cfg := config.Load()
	log.Printf("memotree worker started env=%s concurrency=%d", cfg.AppEnv, cfg.Concurrency)

	// 后续这里接入媒体处理任务队列，避免在 API 请求链路内生成缩略图或视频封面。
	select {}
}
