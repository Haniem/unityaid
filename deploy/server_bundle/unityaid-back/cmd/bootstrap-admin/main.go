package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/bcrypt"

	"unityaid-back/internal/config"
	"unityaid-back/internal/database"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	email := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_EMAIL"))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	firstName := strings.TrimSpace(getEnv("BOOTSTRAP_ADMIN_FIRST_NAME", "Admin"))
	lastName := strings.TrimSpace(getEnv("BOOTSTRAP_ADMIN_LAST_NAME", "Пульс"))
	patronymic := strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_PATRONYMIC"))
	organizationName := strings.TrimSpace(getEnv("BOOTSTRAP_ORGANIZATION_NAME", "Default organization"))
	organizationSlug := strings.TrimSpace(getEnv("BOOTSTRAP_ORGANIZATION_SLUG", "default"))

	if email == "" || password == "" {
		logger.Error("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD are required")
		os.Exit(1)
	}
	if len(password) < 8 {
		logger.Error("BOOTSTRAP_ADMIN_PASSWORD must contain at least 8 characters")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		os.Exit(1)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		logger.Error("failed to start transaction", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var organizationID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, description, contact_email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			contact_email = EXCLUDED.contact_email,
			updated_at = now()
		RETURNING id::text
	`, organizationName, organizationSlug, "Primary client organization", email).Scan(&organizationID); err != nil {
		logger.Error("failed to upsert organization", "error", err)
		os.Exit(1)
	}

	var userID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, patronymic, locale, is_email_verified, is_active)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'ru', true, true)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			patronymic = EXCLUDED.patronymic,
			is_email_verified = true,
			is_active = true,
			updated_at = now()
		RETURNING id::text
	`, email, string(passwordHash), firstName, lastName, patronymic).Scan(&userID); err != nil {
		logger.Error("failed to upsert admin user", "error", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO volunteer_profiles (user_id, city, phone, bio)
		VALUES ($1, '', '', '')
		ON CONFLICT (user_id) DO NOTHING
	`, userID); err != nil {
		logger.Error("failed to ensure volunteer profile", "error", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, status)
		VALUES ($1, $2, 'super_admin', 'active')
		ON CONFLICT (organization_id, user_id) DO UPDATE SET
			role = 'super_admin',
			status = 'active',
			updated_at = now()
	`, organizationID, userID); err != nil {
		logger.Error("failed to ensure organization membership", "error", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO system_roles (code, name, description)
		VALUES ('system_admin', 'Системный администратор', 'Полный доступ ко всей системе, настройкам, пользователям, организациям, справочникам и аудиту.')
		ON CONFLICT (code) DO NOTHING
	`); err != nil {
		logger.Error("failed to ensure system role", "error", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_system_roles (user_id, role_id)
		SELECT $1, id FROM system_roles WHERE code = 'system_admin'
		ON CONFLICT DO NOTHING
	`, userID); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Error("failed to grant system admin role", "error", err)
		}
		os.Exit(1)
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit bootstrap transaction", "error", err)
		os.Exit(1)
	}

	logger.Info("bootstrap admin is ready", "email", email, "organization", organizationSlug)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
