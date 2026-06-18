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
	rows, err := r.db.Query(ctx, invitationSelect()+` ORDER BY ui.created_at DESC`)
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

func (r *Repository) Create(ctx context.Context, email *string, role string, token string, invitedBy string, organizationID *string) (Invitation, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO user_invitations (email, role, token, invited_by, organization_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text
	`, email, role, token, invitedBy, organizationID).Scan(&id)
	if err != nil {
		return Invitation{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) FindByID(ctx context.Context, id string) (Invitation, error) {
	item, err := scanInvitation(r.db.QueryRow(ctx, invitationSelect()+` WHERE ui.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) FindByToken(ctx context.Context, token string) (Invitation, error) {
	item, err := scanInvitation(r.db.QueryRow(ctx, invitationSelect()+` WHERE ui.token = $1`, token))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) AcceptRegistration(ctx context.Context, token string, request AcceptRegistrationRequest, passwordHash string) (Invitation, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Invitation{}, err
	}
	defer tx.Rollback(ctx)

	var invitationID string
	var role string
	var organizationID sql.NullString
	err = tx.QueryRow(ctx, `
		SELECT id::text, role, organization_id::text
		FROM user_invitations
		WHERE token = $1 AND status = 'pending' AND expires_at > now()
		FOR UPDATE
	`, token).Scan(&invitationID, &role, &organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrNotFound
	}
	if err != nil {
		return Invitation{}, err
	}

	var userID string
	err = tx.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, patronymic, locale, is_email_verified, is_active)
		VALUES (lower($1), $2, $3, $4, NULLIF($5, ''), 'ru', true, true)
		RETURNING id::text
	`, request.Email, passwordHash, request.FirstName, request.LastName, request.Patronymic).Scan(&userID)
	if err != nil {
		return Invitation{}, err
	}

	if organizationID.Valid {
		_, err = tx.Exec(ctx, `
			INSERT INTO organization_members (organization_id, user_id, role, status)
			VALUES ($1, $2, $3, 'active')
			ON CONFLICT (organization_id, user_id)
			DO UPDATE SET role = EXCLUDED.role, status = 'active', updated_at = now()
		`, organizationID.String, userID, role)
		if err != nil {
			return Invitation{}, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE user_invitations
		SET status = 'accepted', email = lower($2), accepted_by = $3, accepted_at = now(), updated_at = now()
		WHERE id = $1
	`, invitationID, request.Email, userID)
	if err != nil {
		return Invitation{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}

	return r.FindByID(ctx, invitationID)
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

func invitationSelect() string {
	return `
		SELECT ui.id::text, ui.email, ui.role, ui.token, ui.status, ui.organization_id::text, o.name,
			ui.invited_by::text, ui.accepted_by::text, ui.expires_at, ui.accepted_at, ui.created_at, ui.updated_at
		FROM user_invitations ui
		LEFT JOIN organizations o ON o.id = ui.organization_id
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanInvitation(row scanner) (Invitation, error) {
	var item Invitation
	var email, organizationID, organizationName, invitedBy, acceptedBy sql.NullString
	err := row.Scan(
		&item.ID,
		&email,
		&item.Role,
		&item.Token,
		&item.Status,
		&organizationID,
		&organizationName,
		&invitedBy,
		&acceptedBy,
		&item.ExpiresAt,
		&item.AcceptedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if email.Valid {
		item.Email = &email.String
	}
	if organizationID.Valid {
		item.OrganizationID = &organizationID.String
	}
	if organizationName.Valid {
		item.OrganizationName = &organizationName.String
	}
	if invitedBy.Valid {
		item.InvitedBy = &invitedBy.String
	}
	if acceptedBy.Valid {
		item.AcceptedBy = &acceptedBy.String
	}
	item.Link = "/register?invite=" + item.Token
	return item, err
}
