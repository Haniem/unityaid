package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisURL := getEnv("REDIS_URL", "redis://redis:6379/0")
	queueName := getEnv("WORKER_QUEUE", "default")
	concurrency := getEnvInt("WORKER_CONCURRENCY", 2)

	logger.Info("unityaid worker started", "redis_url", redisURL, "queue", queueName, "concurrency", concurrency)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("unityaid worker stopped")
			return
		case <-ticker.C:
			logger.Info("unityaid worker heartbeat", "queue", queueName)
		}
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
