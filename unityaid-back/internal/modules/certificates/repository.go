package certificates

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("certificate not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, userID string, all bool) ([]Certificate, error) {
	rows, err := r.db.Query(ctx, baseSelect()+`
		WHERE ($1::text = '' OR c.user_id::text = $1 OR $2 = true)
		ORDER BY c.issued_at DESC
	`, userID, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Certificate{}
	for rows.Next() {
		item, err := scanCertificate(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Certificate, error) {
	item, err := scanCertificate(r.db.QueryRow(ctx, baseSelect()+` WHERE c.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) FindByCode(ctx context.Context, code string) (Certificate, error) {
	item, err := scanCertificate(r.db.QueryRow(ctx, baseSelect()+` WHERE c.verify_code = $1`, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return Certificate{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) TotalHours(ctx context.Context, userID string, organizationID *string) (float64, error) {
	var hours float64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(hours), 0)::float8
		FROM time_entries
		WHERE user_id = $1
			AND status = 'approved'
			AND ($2::uuid IS NULL OR organization_id = $2::uuid)
	`, userID, organizationID).Scan(&hours)
	return hours, err
}

func (r *Repository) ParticipationEvents(ctx context.Context, userID string, organizationID *string) ([]ParticipationEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT e.id::text, e.title, o.name, e.starts_at, e.ends_at, e.location
		FROM event_applications ea
		JOIN events e ON e.id = ea.event_id
		LEFT JOIN organizations o ON o.id = e.organization_id
		LEFT JOIN event_attendance att ON att.event_id = e.id AND att.user_id = ea.user_id
		WHERE ea.user_id = $1
			AND (ea.status = 'approved' OR att.id IS NOT NULL)
			AND ($2::uuid IS NULL OR e.organization_id = $2::uuid)
		ORDER BY e.starts_at DESC
	`, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ParticipationEvent{}
	for rows.Next() {
		var item ParticipationEvent
		var organizationName, location sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &organizationName, &item.StartsAt, &item.EndsAt, &location); err != nil {
			return nil, err
		}
		if organizationName.Valid {
			item.OrganizationName = &organizationName.String
		}
		if location.Valid {
			item.Location = &location.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, request GenerateRequest, totalHours float64, verifyCode string, issuedBy string) (Certificate, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO certificates (user_id, organization_id, type, title, description, total_hours, verify_code, issued_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text
	`, request.UserID, request.OrganizationID, normalizeType(request.Type), request.Title, request.Description, totalHours, verifyCode, issuedBy).Scan(&id)
	if err != nil {
		return Certificate{}, err
	}
	return r.FindByID(ctx, id)
}

func baseSelect() string {
	return `
		SELECT c.id::text, c.user_id::text, concat_ws(' ', u.last_name, u.first_name), u.email,
			c.organization_id::text, o.name, c.type, c.title, c.description, c.total_hours::float8,
			c.verify_code, c.issued_by::text, concat_ws(' ', issuer.last_name, issuer.first_name),
			c.issued_at, c.created_at
		FROM certificates c
		JOIN users u ON u.id = c.user_id
		LEFT JOIN organizations o ON o.id = c.organization_id
		LEFT JOIN users issuer ON issuer.id = c.issued_by
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCertificate(row scanner) (Certificate, error) {
	var item Certificate
	var organizationID, organizationName, issuedBy, issuedByName sql.NullString
	err := row.Scan(&item.ID, &item.UserID, &item.UserName, &item.Email, &organizationID, &organizationName, &item.Type, &item.Title, &item.Description, &item.TotalHours, &item.VerifyCode, &issuedBy, &issuedByName, &item.IssuedAt, &item.CreatedAt)
	if organizationID.Valid {
		item.OrganizationID = &organizationID.String
	}
	if organizationName.Valid {
		item.OrganizationName = &organizationName.String
	}
	if issuedBy.Valid {
		item.IssuedBy = &issuedBy.String
	}
	if issuedByName.Valid {
		item.IssuedByName = &issuedByName.String
	}
	return item, err
}

func normalizeType(value string) string {
	switch value {
	case "hours", "participation":
		return value
	default:
		return "participation"
	}
}
