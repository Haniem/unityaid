package audit

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Record(ctx context.Context, entry Entry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_log (
			user_id, method, action, entity_type, entity_id, path, status_code, ip_address, user_agent
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9)
	`, entry.UserID, entry.Method, entry.Action, entry.EntityType, entry.EntityID, entry.Path, entry.StatusCode, entry.IPAddress, entry.UserAgent)
	return err
}
