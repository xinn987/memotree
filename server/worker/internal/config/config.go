package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppEnv      string
	Concurrency int
}

func Load() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", "local"),
		Concurrency: getEnvInt("MEDIA_WORKER_CONCURRENCY", 2),
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
