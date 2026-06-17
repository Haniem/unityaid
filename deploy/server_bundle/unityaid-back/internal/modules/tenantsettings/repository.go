package tenantsettings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Get(ctx context.Context) (TenantSettings, error) {
	item, err := scanSettings(r.db.QueryRow(ctx, `
		SELECT id::text, display_name, description, logo_url, primary_color, accent_color, timezone, locale,
			contact_email, contact_phone, default_organization_id::text, default_organization_name,
			default_organization_slug, onboarding_completed, pending_invites, created_at, updated_at
		FROM tenant_settings
		WHERE singleton_key = true
	`))
	if errors.Is(err, pgx.ErrNoRows) {
		if _, createErr := r.db.Exec(ctx, "INSERT INTO tenant_settings (singleton_key) VALUES (true) ON CONFLICT (singleton_key) DO NOTHING"); createErr != nil {
			return TenantSettings{}, createErr
		}
		return r.Get(ctx)
	}
	return item, err
}

func (r *Repository) Upsert(ctx context.Context, request UpsertRequest) (TenantSettings, error) {
	invites, err := json.Marshal(request.PendingInvites)
	if err != nil {
		return TenantSettings{}, err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO tenant_settings (
			singleton_key, display_name, description, logo_url, primary_color, accent_color, timezone, locale,
			contact_email, contact_phone, default_organization_id, default_organization_name,
			default_organization_slug, onboarding_completed, pending_invites
		)
		VALUES (true, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (singleton_key)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			logo_url = EXCLUDED.logo_url,
			primary_color = EXCLUDED.primary_color,
			accent_color = EXCLUDED.accent_color,
			timezone = EXCLUDED.timezone,
			locale = EXCLUDED.locale,
			contact_email = EXCLUDED.contact_email,
			contact_phone = EXCLUDED.contact_phone,
			default_organization_id = EXCLUDED.default_organization_id,
			default_organization_name = EXCLUDED.default_organization_name,
			default_organization_slug = EXCLUDED.default_organization_slug,
			onboarding_completed = EXCLUDED.onboarding_completed,
			pending_invites = EXCLUDED.pending_invites,
			updated_at = now()
	`, request.DisplayName, request.Description, request.LogoURL, request.PrimaryColor, request.AccentColor,
		request.Timezone, request.Locale, request.ContactEmail, request.ContactPhone, request.DefaultOrganizationID,
		request.DefaultOrganizationName, request.DefaultOrganizationSlug, request.OnboardingCompleted, invites)
	if err != nil {
		return TenantSettings{}, err
	}
	return r.Get(ctx)
}

func (r *Repository) CreateDefaultOrganization(ctx context.Context, name string, slug string, description string, contactEmail *string) (string, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, description, contact_email)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, name, slug, description, contactEmail).Scan(&id)
	return id, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSettings(row scanner) (TenantSettings, error) {
	var item TenantSettings
	var logoURL, contactEmail, contactPhone, organizationID sql.NullString
	var invites []byte
	err := row.Scan(
		&item.ID,
		&item.DisplayName,
		&item.Description,
		&logoURL,
		&item.PrimaryColor,
		&item.AccentColor,
		&item.Timezone,
		&item.Locale,
		&contactEmail,
		&contactPhone,
		&organizationID,
		&item.DefaultOrganizationName,
		&item.DefaultOrganizationSlug,
		&item.OnboardingCompleted,
		&invites,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if logoURL.Valid {
		item.LogoURL = &logoURL.String
	}
	if contactEmail.Valid {
		item.ContactEmail = &contactEmail.String
	}
	if contactPhone.Valid {
		item.ContactPhone = &contactPhone.String
	}
	if organizationID.Valid {
		item.DefaultOrganizationID = &organizationID.String
	}
	if len(invites) > 0 {
		_ = json.Unmarshal(invites, &item.PendingInvites)
	}
	if item.PendingInvites == nil {
		item.PendingInvites = []PendingInvite{}
	}
	return item, err
}
