package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"unityaid-back/internal/config"
	"unityaid-back/internal/database"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunSeeds(ctx, db, "seeds"); err != nil {
		logger.Error("failed to run seeds", "error", err)
		os.Exit(1)
	}

	logger.Info("seeds applied")
}
