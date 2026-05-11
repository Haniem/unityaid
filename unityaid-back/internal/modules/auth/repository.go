package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")
var ErrTokenNotFound = errors.New("token not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	user, err := r.findOne(ctx, `
		SELECT id::text, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_email_verified, is_active, last_login_at
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
		SELECT id::text, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_email_verified, is_active, last_login_at
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

func (r *Repository) CreateUser(ctx context.Context, request RegisterRequest, passwordHash string) (User, error) {
	var user User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, locale, is_email_verified)
		VALUES (lower($1), $2, $3, $4, 'ru', false)
		RETURNING id::text, email, password_hash, first_name, last_name, patronymic, avatar_url, locale, is_email_verified, is_active, last_login_at
	`, request.Email, passwordHash, request.FirstName, request.LastName).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Patronymic,
		&user.AvatarURL,
		&user.Locale,
		&user.IsEmailVerified,
		&user.IsActive,
		&user.LastLoginAt,
	)
	if err != nil {
		return User{}, err
	}

	if err := r.loadMemberships(ctx, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1", userID, passwordHash)
	return err
}

func (r *Repository) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET is_email_verified = true, updated_at = now() WHERE id = $1", userID)
	return err
}

func (r *Repository) CreateRefreshSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time, userAgent string, ipAddress string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, tokenHash, expiresAt, userAgent, ipAddress)
	return err
}

func (r *Repository) FindUserByRefreshTokenHash(ctx context.Context, tokenHash string) (User, error) {
	var userID string
	err := r.db.QueryRow(ctx, `
		SELECT user_id::text
		FROM refresh_sessions
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > now()
	`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrTokenNotFound
	}
	if err != nil {
		return User{}, err
	}
	return r.FindByID(ctx, userID)
}

func (r *Repository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL", tokenHash)
	return err
}

func (r *Repository) RevokeUserRefreshSessions(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, "UPDATE refresh_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL", userID)
	return err
}

func (r *Repository) RevokeAccessToken(ctx context.Context, jti string, userID string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO revoked_access_tokens (jti, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (jti) DO NOTHING
	`, jti, userID, expiresAt)
	return err
}

func (r *Repository) IsAccessTokenRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}
	var revoked bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM revoked_access_tokens
			WHERE jti = $1 AND expires_at > now()
		)
	`, jti).Scan(&revoked)
	return revoked, err
}

func (r *Repository) HasAnyRole(ctx context.Context, userID string, roles ...string) (bool, error) {
	if userID == "" || len(roles) == 0 {
		return false, nil
	}

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM organization_members
			WHERE user_id = $1
				AND status = 'active'
				AND role::text = ANY($2::text[])
		)
	`, userID, roles).Scan(&exists)
	return exists, err
}

func (r *Repository) HasRoleInOrganization(ctx context.Context, userID string, organizationID string, roles ...string) (bool, error) {
	if userID == "" || organizationID == "" || len(roles) == 0 {
		return false, nil
	}

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM organization_members
			WHERE user_id = $1
				AND organization_id = $2
				AND status = 'active'
				AND role::text = ANY($3::text[])
		)
	`, userID, organizationID, roles).Scan(&exists)
	return exists, err
}

func (r *Repository) SharesOrganizationWithRole(ctx context.Context, actorUserID string, targetUserID string, roles ...string) (bool, error) {
	if actorUserID == "" || targetUserID == "" || len(roles) == 0 {
		return false, nil
	}

	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM organization_members actor
			JOIN organization_members target ON target.organization_id = actor.organization_id
			WHERE actor.user_id = $1
				AND target.user_id = $2
				AND actor.status = 'active'
				AND target.status = 'active'
				AND actor.role::text = ANY($3::text[])
		)
	`, actorUserID, targetUserID, roles).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateEmailVerificationToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) ConsumeEmailVerificationToken(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, `
		UPDATE email_verification_tokens
		SET used_at = now()
		WHERE token_hash = $1
			AND used_at IS NULL
			AND expires_at > now()
		RETURNING user_id::text
	`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrTokenNotFound
	}
	return userID, err
}

func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *Repository) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1
			AND used_at IS NULL
			AND expires_at > now()
		RETURNING user_id::text
	`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrTokenNotFound
	}
	return userID, err
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
		&user.IsEmailVerified,
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
