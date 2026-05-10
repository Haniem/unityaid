package config

import (
	"os"
	"strings"
)

type Config struct {
	AppEnv             string
	HTTPPort          string
	DatabaseURL       string
	CORSAllowedOrigins []string
	JWTSecret          string
	RunSeeds           bool
	UploadsDir         string
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://unityaid:unityaid@localhost:5432/unityaid?sslmode=disable"),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		JWTSecret:          getEnv("JWT_SECRET", "dev-change-me"),
		RunSeeds:           getEnv("RUN_SEEDS", "true") == "true",
		UploadsDir:         getEnv("UPLOADS_DIR", "uploads"),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
