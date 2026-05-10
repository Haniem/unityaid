package organizations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("organization not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Organization, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, name, slug, description, contact_email, created_at, updated_at FROM organizations ORDER BY name`)
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
	item, err := scanOrganization(r.db.QueryRow(ctx, `SELECT id::text, name, slug, description, contact_email, created_at, updated_at FROM organizations WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest) (Organization, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, description, contact_email)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, request.Name, request.Slug, request.Description, request.ContactEmail).Scan(&id)
	if err != nil {
		return Organization{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest) (Organization, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE organizations
		SET name = $2, slug = $3, description = $4, contact_email = $5, updated_at = now()
		WHERE id = $1
	`, id, request.Name, request.Slug, request.Description, request.ContactEmail)
	if err != nil {
		return Organization{}, err
	}
	if tag.RowsAffected() == 0 {
		return Organization{}, ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOrganization(row scanner) (Organization, error) {
	var item Organization
	var email sql.NullString
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &email, &item.CreatedAt, &item.UpdatedAt)
	if email.Valid {
		item.ContactEmail = &email.String
	}
	return item, err
}
