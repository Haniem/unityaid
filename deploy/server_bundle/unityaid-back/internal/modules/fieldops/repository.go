package fieldops

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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

func (r *Repository) CreateQRToken(ctx context.Context, eventID string, request CreateQRRequest, createdBy string) (QRToken, error) {
	if request.Mode == "" {
		request.Mode = "attendance"
	}
	token, err := randomToken()
	if err != nil {
		return QRToken{}, err
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO field_qr_tokens (event_id, token, mode, expires_at, created_by)
		VALUES ($1, $2, $3, NULLIF($4, '')::timestamptz, NULLIF($5, '')::uuid)
		RETURNING id::text, event_id::text, token, mode, COALESCE(to_char(expires_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''), to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	`, eventID, token, request.Mode, request.ExpiresAt, createdBy)
	var item QRToken
	err = row.Scan(&item.ID, &item.EventID, &item.Token, &item.Mode, &item.ExpiresAt, &item.CreatedAt)
	return item, err
}

func (r *Repository) Scan(ctx context.Context, token string, userID string) (Checkin, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Checkin{}, err
	}
	defer tx.Rollback(ctx)

	var tokenID, eventID, mode string
	var expired bool
	err = tx.QueryRow(ctx, `
		SELECT id::text, event_id::text, mode, expires_at IS NOT NULL AND expires_at < now()
		FROM field_qr_tokens
		WHERE token = $1
	`, token).Scan(&tokenID, &eventID, &mode, &expired)
	if errors.Is(err, pgx.ErrNoRows) {
		return Checkin{}, ErrNotFound
	}
	if err != nil {
		return Checkin{}, err
	}
	if expired {
		return Checkin{}, ErrExpired
	}

	var item Checkin
	if mode == "checkout" {
		err = tx.QueryRow(ctx, `
			UPDATE field_checkins
			SET checkout_at = now(), status = 'checked_out', qr_token_id = $3, updated_at = now()
			WHERE event_id = $1 AND user_id = $2
			RETURNING id::text, event_id::text, user_id::text, COALESCE(qr_token_id::text, ''), COALESCE(to_char(checkin_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''), COALESCE(to_char(checkout_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''), status, source
		`, eventID, userID, tokenID).Scan(&item.ID, &item.EventID, &item.UserID, &item.QRTokenID, &item.CheckinAt, &item.CheckoutAt, &item.Status, &item.Source)
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO field_checkins (event_id, user_id, qr_token_id, checkin_at, status)
			VALUES ($1, $2, $3, now(), 'checked_in')
			ON CONFLICT (event_id, user_id) DO UPDATE SET
				qr_token_id = EXCLUDED.qr_token_id,
				checkin_at = COALESCE(field_checkins.checkin_at, now()),
				status = CASE WHEN field_checkins.checkout_at IS NULL THEN 'checked_in' ELSE field_checkins.status END,
				updated_at = now()
			RETURNING id::text, event_id::text, user_id::text, COALESCE(qr_token_id::text, ''), COALESCE(to_char(checkin_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''), COALESCE(to_char(checkout_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''), status, source
		`, eventID, userID, tokenID).Scan(&item.ID, &item.EventID, &item.UserID, &item.QRTokenID, &item.CheckinAt, &item.CheckoutAt, &item.Status, &item.Source)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Checkin{}, ErrNotFound
	}
	if err != nil {
		return Checkin{}, err
	}
	if mode == "checkout" {
		_, err = tx.Exec(ctx, `
			INSERT INTO time_entries (organization_id, user_id, event_id, hours, description, status, reviewed_by, reviewed_at)
			SELECT e.organization_id, fc.user_id, fc.event_id,
				GREATEST(ROUND((EXTRACT(EPOCH FROM (fc.checkout_at - fc.checkin_at)) / 3600)::numeric, 2), 0.01),
				'QR field check-in ' || fc.id::text,
				'approved',
				fqt.created_by,
				now()
			FROM field_checkins fc
			JOIN events e ON e.id = fc.event_id
			LEFT JOIN field_qr_tokens fqt ON fqt.id = fc.qr_token_id
			WHERE fc.id = $1
				AND fc.checkin_at IS NOT NULL
				AND fc.checkout_at IS NOT NULL
				AND NOT EXISTS (
					SELECT 1 FROM time_entries te
					WHERE te.event_id = fc.event_id
						AND te.user_id = fc.user_id
						AND te.description = 'QR field check-in ' || fc.id::text
				)
		`, item.ID)
		if err != nil {
			return Checkin{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Checkin{}, err
	}
	return item, nil
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
