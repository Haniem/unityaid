package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	user, err := r.findOne(ctx, `
		SELECT id::text, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_active, last_login_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)
	if err != nil {
		return User{}, err
	}

	if err := r.loadMemberships(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (User, error) {
	user, err := r.findOne(ctx, `
		SELECT id::text, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_active, last_login_at
		FROM users
		WHERE id = $1
	`, id)
	if err != nil {
		return User{}, err
	}

	if err := r.loadMemberships(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1", userID)
	return err
}

func (r *Repository) findOne(ctx context.Context, query string, args ...any) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
		&user.AvatarURL,
		&user.Locale,
		&user.IsActive,
		&user.LastLoginAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *Repository) loadMemberships(ctx context.Context, user *User) error {
	rows, err := r.db.Query(ctx, `
		SELECT om.organization_id::text, o.name, om.role::text, om.status::text
		FROM organization_members om
		JOIN organizations o ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY
			CASE om.role
				WHEN 'super_admin' THEN 1
				WHEN 'org_admin' THEN 2
				WHEN 'coordinator' THEN 3
				ELSE 4
			END,
			o.name
	`, user.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	user.Organizations = []Membership{}
	for rows.Next() {
		var membership Membership
		if err := rows.Scan(&membership.OrganizationID, &membership.OrganizationName, &membership.Role, &membership.Status); err != nil {
			return err
		}
		user.Organizations = append(user.Organizations, membership)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(user.Organizations) > 0 {
		user.PrimaryRole = user.Organizations[0].Role
		user.OrganizationID = &user.Organizations[0].OrganizationID
	}

	return nil
}
