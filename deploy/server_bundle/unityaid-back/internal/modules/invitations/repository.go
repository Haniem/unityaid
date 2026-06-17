package invitations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("invitation not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Invitation, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, email, role, token, status, invited_by::text, accepted_by::text,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invitations
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Invitation{}
	for rows.Next() {
		item, err := scanInvitation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, email string, role string, token string, invitedBy string) (Invitation, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_invitations (email, role, token, invited_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, email, role, token, invitedBy).Scan(&id)
	if err != nil {
		return Invitation{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) FindByID(ctx context.Context, id string) (Invitation, error) {
	item, err := scanInvitation(r.db.QueryRow(ctx, `
		SELECT id::text, email, role, token, status, invited_by::text, accepted_by::text,
			expires_at, accepted_at, created_at, updated_at
		FROM user_invitations
		WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Accept(ctx context.Context, token string, userID string) (Invitation, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		UPDATE user_invitations
		SET status = 'accepted', accepted_by = $2, accepted_at = now(), updated_at = now()
		WHERE token = $1 AND status = 'pending' AND expires_at > now()
		RETURNING id::text
	`, token, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}
	if err != nil {
		return Invitation{}, err
	}
	return r.FindByID(ctx, id)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanInvitation(row scanner) (Invitation, error) {
	var item Invitation
	var invitedBy, acceptedBy sql.NullString
	err := row.Scan(&item.ID, &item.Email, &item.Role, &item.Token, &item.Status, &invitedBy, &acceptedBy, &item.ExpiresAt, &item.AcceptedAt, &item.CreatedAt, &item.UpdatedAt)
	if invitedBy.Valid {
		item.InvitedBy = &invitedBy.String
	}
	if acceptedBy.Valid {
		item.AcceptedBy = &acceptedBy.String
	}
	item.Link = "/register?invite=" + item.Token
	return item, err
}
