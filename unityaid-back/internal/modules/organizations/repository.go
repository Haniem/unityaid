package organizations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("organization not found")
var ErrMemberNotFound = errors.New("organization member not found")
var ErrUserNotFound = errors.New("user not found")
var ErrHasImportantData = errors.New("organization has important linked data")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]Organization, error) {
	query := `
		SELECT id::text, name, slug, description, contact_email, logo_url, website_url, phone, address,
			deleted_at IS NOT NULL, created_at, updated_at
		FROM organizations
		WHERE ($1 = true OR deleted_at IS NULL)
			AND (
				$2 = ''
				OR name ILIKE '%' || $2 || '%'
				OR slug ILIKE '%' || $2 || '%'
				OR description ILIKE '%' || $2 || '%'
				OR contact_email ILIKE '%' || $2 || '%'
			)
		ORDER BY deleted_at NULLS FIRST, name
	`
	rows, err := r.db.Query(ctx, query, filters.IncludeDeleted, filters.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Organization{}
	for rows.Next() {
		item, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Organization, error) {
	item, err := scanOrganization(r.db.QueryRow(ctx, `
		SELECT id::text, name, slug, description, contact_email, logo_url, website_url, phone, address,
			deleted_at IS NOT NULL, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest) (Organization, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, description, contact_email, logo_url, website_url, phone, address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text
	`, request.Name, request.Slug, request.Description, request.ContactEmail, request.LogoURL, request.WebsiteURL, request.Phone, request.Address).Scan(&id)
	if err != nil {
		return Organization{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest) (Organization, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE organizations
		SET name = $2,
			slug = $3,
			description = $4,
			contact_email = $5,
			logo_url = $6,
			website_url = $7,
			phone = $8,
			address = $9,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, request.Name, request.Slug, request.Description, request.ContactEmail, request.LogoURL, request.WebsiteURL, request.Phone, request.Address)
	if err != nil {
		return Organization{}, err
	}
	if tag.RowsAffected() == 0 {
		return Organization{}, ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE organizations
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) HasImportantData(ctx context.Context, id string) (bool, error) {
	var hasData bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM events WHERE organization_id = $1)
			OR EXISTS (SELECT 1 FROM tasks WHERE organization_id = $1)
			OR EXISTS (SELECT 1 FROM news WHERE organization_id = $1)
	`, id).Scan(&hasData)
	return hasData, err
}

func (r *Repository) ListMembers(ctx context.Context, organizationID string) ([]OrganizationMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT om.id::text, u.id::text, u.email, u.first_name, u.last_name, u.avatar_url,
			om.role::text, om.status::text, om.created_at, om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1
		ORDER BY
			CASE om.role
				WHEN 'super_admin' THEN 1
				WHEN 'org_admin' THEN 2
				WHEN 'coordinator' THEN 3
				ELSE 4
			END,
			u.last_name,
			u.first_name
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []OrganizationMember{}
	for rows.Next() {
		item, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) AddMember(ctx context.Context, organizationID string, request AddMemberRequest) (OrganizationMember, error) {
	userID, err := r.findUserIDByEmail(ctx, request.Email)
	if err != nil {
		return OrganizationMember{}, err
	}

	var memberID string
	err = r.db.QueryRow(ctx, `
		INSERT INTO organization_members (organization_id, user_id, role, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (organization_id, user_id)
		DO UPDATE SET role = EXCLUDED.role, status = EXCLUDED.status, updated_at = now()
		RETURNING id::text
	`, organizationID, userID, request.Role, request.Status).Scan(&memberID)
	if err != nil {
		return OrganizationMember{}, err
	}
	return r.FindMember(ctx, organizationID, memberID)
}

func (r *Repository) UpdateMember(ctx context.Context, organizationID string, memberID string, request UpdateMemberRequest) (OrganizationMember, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE organization_members
		SET role = $3, status = $4, updated_at = now()
		WHERE organization_id = $1 AND id = $2
	`, organizationID, memberID, request.Role, request.Status)
	if err != nil {
		return OrganizationMember{}, err
	}
	if tag.RowsAffected() == 0 {
		return OrganizationMember{}, ErrMemberNotFound
	}
	return r.FindMember(ctx, organizationID, memberID)
}

func (r *Repository) DeleteMember(ctx context.Context, organizationID string, memberID string) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND id = $2
	`, organizationID, memberID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *Repository) FindMember(ctx context.Context, organizationID string, memberID string) (OrganizationMember, error) {
	item, err := scanMember(r.db.QueryRow(ctx, `
		SELECT om.id::text, u.id::text, u.email, u.first_name, u.last_name, u.avatar_url,
			om.role::text, om.status::text, om.created_at, om.updated_at
		FROM organization_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.organization_id = $1 AND om.id = $2
	`, organizationID, memberID))
	if errors.Is(err, pgx.ErrNoRows) {
		return OrganizationMember{}, ErrMemberNotFound
	}
	return item, err
}

func (r *Repository) findUserIDByEmail(ctx context.Context, email string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, "SELECT id::text FROM users WHERE lower(email) = lower($1)", email).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	return userID, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOrganization(row scanner) (Organization, error) {
	var item Organization
	var email, logoURL, websiteURL, phone, address sql.NullString
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Slug,
		&item.Description,
		&email,
		&logoURL,
		&websiteURL,
		&phone,
		&address,
		&item.IsDeleted,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if email.Valid {
		item.ContactEmail = &email.String
	}
	if logoURL.Valid {
		item.LogoURL = &logoURL.String
	}
	if websiteURL.Valid {
		item.WebsiteURL = &websiteURL.String
	}
	if phone.Valid {
		item.Phone = &phone.String
	}
	if address.Valid {
		item.Address = &address.String
	}
	return item, err
}

func scanMember(row scanner) (OrganizationMember, error) {
	var item OrganizationMember
	var avatarURL sql.NullString
	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Email,
		&item.FirstName,
		&item.LastName,
		&avatarURL,
		&item.Role,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if avatarURL.Valid {
		item.AvatarURL = &avatarURL.String
	}
	return item, err
}
