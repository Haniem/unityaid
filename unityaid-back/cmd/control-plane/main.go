package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"unityaid-back/internal/database"
)

type app struct {
	db     *pgxpool.Pool
	apiKey string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	port := getEnv("CONTROL_PLANE_PORT", "8090")
	databaseURL := getEnv("CONTROL_PLANE_DATABASE_URL", os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		logger.Error("CONTROL_PLANE_DATABASE_URL or DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect to control plane database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := ensureSchema(ctx, db); err != nil {
		logger.Error("failed to ensure control plane schema", "error", err)
		os.Exit(1)
	}

	service := &app{db: db, apiKey: os.Getenv("CONTROL_PLANE_API_KEY")}
	if service.apiKey == "" {
		logger.Warn("CONTROL_PLANE_API_KEY is empty; protected endpoints are open")
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "unityaid-control-plane", "status": "ok"})
	})

	api := router.Group("/api/v1", service.requireAPIKey)
	api.GET("/plans", service.listPlans)
	api.GET("/clients", service.listClients)
	api.POST("/clients", service.createClient)
	api.GET("/clients/:id", service.getClient)
	api.PATCH("/clients/:id", service.updateClient)
	api.GET("/clients/:id/environments", service.listEnvironments)
	api.POST("/clients/:id/environments", service.createEnvironment)
	api.POST("/clients/:id/deployments", service.createDeployment)
	api.GET("/clients/:id/deployments", service.listDeployments)
	api.POST("/clients/:id/backups", service.createBackup)
	api.GET("/clients/:id/backups", service.listBackups)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("control plane started", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("control plane stopped with error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func (a *app) requireAPIKey(c *gin.Context) {
	if a.apiKey == "" {
		c.Next()
		return
	}
	token := strings.TrimSpace(c.GetHeader("X-Control-Plane-Key"))
	if token == "" {
		auth := c.GetHeader("Authorization")
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	if token != a.apiKey {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Invalid control plane key"})
		return
	}
	c.Next()
}

func (a *app) listPlans(c *gin.Context) {
	rows, err := a.db.Query(c.Request.Context(), `
		SELECT id::text, code, name, monthly_price, limits, features
		FROM cp_plans
		WHERE is_active = true
		ORDER BY monthly_price ASC, code ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	defer rows.Close()

	items, err := collectRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (a *app) listClients(c *gin.Context) {
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)
	search := strings.TrimSpace(c.Query("search"))

	rows, err := a.db.Query(c.Request.Context(), `
		SELECT id::text, slug, name, status, plan_code, primary_domain, owner_email, created_at, updated_at
		FROM cp_clients
		WHERE $1 = '' OR slug ILIKE '%' || $1 || '%' OR name ILIKE '%' || $1 || '%' OR owner_email ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	defer rows.Close()

	items, err := collectRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "limit": limit, "offset": offset})
}

func (a *app) createClient(c *gin.Context) {
	var request clientRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	request.normalize()
	if request.Slug == "" || request.Name == "" || request.OwnerEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "slug, name and ownerEmail are required"})
		return
	}
	if request.PlanCode == "" {
		request.PlanCode = "start"
	}
	if request.Status == "" {
		request.Status = "provisioning"
	}

	item, err := queryOne(c.Request.Context(), a.db, `
		INSERT INTO cp_clients (slug, name, status, plan_code, primary_domain, owner_email, modules, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, slug, name, status, plan_code, primary_domain, owner_email, modules, notes, created_at, updated_at
	`, request.Slug, request.Name, request.Status, request.PlanCode, request.PrimaryDomain, request.OwnerEmail, request.Modules, request.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (a *app) getClient(c *gin.Context) {
	item, err := a.loadClient(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err == pgx.ErrNoRows {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "client_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (a *app) updateClient(c *gin.Context) {
	var request clientRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	request.normalize()
	item, err := queryOne(c.Request.Context(), a.db, `
		UPDATE cp_clients
		SET
			name = COALESCE(NULLIF($2, ''), name),
			status = COALESCE(NULLIF($3, ''), status),
			plan_code = COALESCE(NULLIF($4, ''), plan_code),
			primary_domain = COALESCE(NULLIF($5, ''), primary_domain),
			owner_email = COALESCE(NULLIF($6, ''), owner_email),
			modules = COALESCE($7, modules),
			notes = COALESCE(NULLIF($8, ''), notes),
			updated_at = now()
		WHERE id = $1
		RETURNING id::text, slug, name, status, plan_code, primary_domain, owner_email, modules, notes, created_at, updated_at
	`, c.Param("id"), request.Name, request.Status, request.PlanCode, request.PrimaryDomain, request.OwnerEmail, request.Modules, request.Notes)
	if err != nil {
		status := http.StatusInternalServerError
		if err == pgx.ErrNoRows {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "client_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (a *app) listEnvironments(c *gin.Context) {
	a.listChildRows(c, `
		SELECT id::text, client_id::text, name, kind, status, stack_path, app_version, frontend_url, backend_url, health_status, last_health_at, created_at, updated_at
		FROM cp_environments
		WHERE client_id = $1
		ORDER BY created_at DESC
	`)
}

func (a *app) createEnvironment(c *gin.Context) {
	var request environmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	if request.Name == "" {
		request.Name = "production"
	}
	if request.Kind == "" {
		request.Kind = "compose"
	}
	if request.Status == "" {
		request.Status = "planned"
	}
	item, err := queryOne(c.Request.Context(), a.db, `
		INSERT INTO cp_environments (client_id, name, kind, status, stack_path, app_version, frontend_url, backend_url, health_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'unknown')
		RETURNING id::text, client_id::text, name, kind, status, stack_path, app_version, frontend_url, backend_url, health_status, last_health_at, created_at, updated_at
	`, c.Param("id"), request.Name, request.Kind, request.Status, request.StackPath, request.AppVersion, request.FrontendURL, request.BackendURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (a *app) listDeployments(c *gin.Context) {
	a.listChildRows(c, `
		SELECT id::text, client_id::text, environment_id::text, version, status, triggered_by, log, started_at, finished_at
		FROM cp_deployments
		WHERE client_id = $1
		ORDER BY started_at DESC
	`)
}

func (a *app) createDeployment(c *gin.Context) {
	var request deploymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	if request.Status == "" {
		request.Status = "planned"
	}
	item, err := queryOne(c.Request.Context(), a.db, `
		INSERT INTO cp_deployments (client_id, environment_id, version, status, triggered_by, log)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6)
		RETURNING id::text, client_id::text, environment_id::text, version, status, triggered_by, log, started_at, finished_at
	`, c.Param("id"), request.EnvironmentID, request.Version, request.Status, request.TriggeredBy, request.Log)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (a *app) listBackups(c *gin.Context) {
	a.listChildRows(c, `
		SELECT id::text, client_id::text, environment_id::text, status, backup_path, size_bytes, created_by, created_at, restored_at
		FROM cp_backups
		WHERE client_id = $1
		ORDER BY created_at DESC
	`)
}

func (a *app) createBackup(c *gin.Context) {
	var request backupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	if request.Status == "" {
		request.Status = "planned"
	}
	item, err := queryOne(c.Request.Context(), a.db, `
		INSERT INTO cp_backups (client_id, environment_id, status, backup_path, size_bytes, created_by)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6)
		RETURNING id::text, client_id::text, environment_id::text, status, backup_path, size_bytes, created_by, created_at, restored_at
	`, c.Param("id"), request.EnvironmentID, request.Status, request.BackupPath, request.SizeBytes, request.CreatedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (a *app) loadClient(ctx context.Context, id string) (map[string]any, error) {
	return queryOne(ctx, a.db, `
		SELECT id::text, slug, name, status, plan_code, primary_domain, owner_email, modules, notes, created_at, updated_at
		FROM cp_clients
		WHERE id = $1 OR slug = $1
	`, id)
}

func (a *app) listChildRows(c *gin.Context, sql string) {
	rows, err := a.db.Query(c.Request.Context(), sql, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	defer rows.Close()
	items, err := collectRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type clientRequest struct {
	Slug          string   `json:"slug"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	PlanCode      string   `json:"planCode"`
	PrimaryDomain string   `json:"primaryDomain"`
	OwnerEmail    string   `json:"ownerEmail"`
	Modules       []string `json:"modules"`
	Notes         string   `json:"notes"`
}

func (r *clientRequest) normalize() {
	r.Slug = strings.TrimSpace(strings.ToLower(r.Slug))
	r.Name = strings.TrimSpace(r.Name)
	r.Status = strings.TrimSpace(strings.ToLower(r.Status))
	r.PlanCode = strings.TrimSpace(strings.ToLower(r.PlanCode))
	r.PrimaryDomain = strings.TrimSpace(strings.ToLower(r.PrimaryDomain))
	r.OwnerEmail = strings.TrimSpace(strings.ToLower(r.OwnerEmail))
}

type environmentRequest struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	StackPath   string `json:"stackPath"`
	AppVersion  string `json:"appVersion"`
	FrontendURL string `json:"frontendUrl"`
	BackendURL  string `json:"backendUrl"`
}

type deploymentRequest struct {
	EnvironmentID string `json:"environmentId"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	TriggeredBy   string `json:"triggeredBy"`
	Log           string `json:"log"`
}

type backupRequest struct {
	EnvironmentID string `json:"environmentId"`
	Status        string `json:"status"`
	BackupPath    string `json:"backupPath"`
	SizeBytes     int64  `json:"sizeBytes"`
	CreatedBy     string `json:"createdBy"`
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil || value < 0 {
		return fallback
	}
	if value > 200 {
		return 200
	}
	return value
}

func collectRows(rows pgx.Rows) ([]map[string]any, error) {
	items := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		fields := rows.FieldDescriptions()
		item := make(map[string]any, len(fields))
		for i, field := range fields {
			item[string(field.Name)] = values[i]
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func queryOne(ctx context.Context, db *pgxpool.Pool, sql string, args ...any) (map[string]any, error) {
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := collectRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, pgx.ErrNoRows
	}
	return items[0], nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func ensureSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS cp_plans (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	code TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	monthly_price INTEGER NOT NULL DEFAULT 0,
	limits JSONB NOT NULL DEFAULT '{}'::jsonb,
	features JSONB NOT NULL DEFAULT '[]'::jsonb,
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cp_clients (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'provisioning',
	plan_code TEXT NOT NULL DEFAULT 'start',
	primary_domain TEXT,
	owner_email TEXT NOT NULL,
	modules TEXT[] NOT NULL DEFAULT '{}',
	notes TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cp_environments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	client_id UUID NOT NULL REFERENCES cp_clients(id) ON DELETE CASCADE,
	name TEXT NOT NULL DEFAULT 'production',
	kind TEXT NOT NULL DEFAULT 'compose',
	status TEXT NOT NULL DEFAULT 'planned',
	stack_path TEXT NOT NULL DEFAULT '',
	app_version TEXT NOT NULL DEFAULT '',
	frontend_url TEXT NOT NULL DEFAULT '',
	backend_url TEXT NOT NULL DEFAULT '',
	health_status TEXT NOT NULL DEFAULT 'unknown',
	last_health_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cp_deployments (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	client_id UUID NOT NULL REFERENCES cp_clients(id) ON DELETE CASCADE,
	environment_id UUID REFERENCES cp_environments(id) ON DELETE SET NULL,
	version TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL DEFAULT 'planned',
	triggered_by TEXT NOT NULL DEFAULT '',
	log TEXT NOT NULL DEFAULT '',
	started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cp_backups (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	client_id UUID NOT NULL REFERENCES cp_clients(id) ON DELETE CASCADE,
	environment_id UUID REFERENCES cp_environments(id) ON DELETE SET NULL,
	status TEXT NOT NULL DEFAULT 'planned',
	backup_path TEXT NOT NULL DEFAULT '',
	size_bytes BIGINT NOT NULL DEFAULT 0,
	created_by TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	restored_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cp_feature_flags (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	client_id UUID NOT NULL REFERENCES cp_clients(id) ON DELETE CASCADE,
	code TEXT NOT NULL,
	is_enabled BOOLEAN NOT NULL DEFAULT false,
	config JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	UNIQUE (client_id, code)
);

CREATE INDEX IF NOT EXISTS idx_cp_clients_status ON cp_clients(status);
CREATE INDEX IF NOT EXISTS idx_cp_environments_client_id ON cp_environments(client_id);
CREATE INDEX IF NOT EXISTS idx_cp_deployments_client_id ON cp_deployments(client_id);
CREATE INDEX IF NOT EXISTS idx_cp_backups_client_id ON cp_backups(client_id);

INSERT INTO cp_plans (code, name, monthly_price, limits, features)
VALUES
	('start', 'Start', 0, '{"volunteers":100,"organizations":1}'::jsonb, '["events","tasks","time_entries","certificates"]'::jsonb),
	('organization', 'Organization', 9900, '{"volunteers":1000,"organizations":10}'::jsonb, '["events","tasks","time_entries","analytics","audit","certificates","knowledge_base"]'::jsonb),
	('network', 'Network', 29900, '{"volunteers":10000,"organizations":100}'::jsonb, '["events","tasks","time_entries","analytics","audit","certificates","knowledge_base","custom_domain"]'::jsonb),
	('enterprise', 'Enterprise', 0, '{}'::jsonb, '["all","on_prem","sso"]'::jsonb)
ON CONFLICT (code) DO UPDATE SET
	name = EXCLUDED.name,
	monthly_price = EXCLUDED.monthly_price,
	limits = EXCLUDED.limits,
	features = EXCLUDED.features;
`)
	return err
}
