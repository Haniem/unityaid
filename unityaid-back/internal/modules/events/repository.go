package events

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("event not found")
var ErrAlreadyExists = errors.New("event application already exists")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Event, error) {
	rows, err := r.db.Query(ctx, baseSelect()+` ORDER BY e.starts_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Event{}
	for rows.Next() {
		item, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Event, error) {
	item, err := scanEvent(r.db.QueryRow(ctx, baseSelect()+` WHERE e.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Event{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest, startsAt time.Time, endsAt time.Time, userID string) (Event, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO events (organization_id, title, description, format, status, starts_at, ends_at, location, max_participants, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text
	`, request.OrganizationID, request.Title, request.Description, normalizeFormat(request.Format), normalizeStatus(request.Status), startsAt, endsAt, request.Location, request.MaxParticipants, userID).Scan(&id)
	if err != nil {
		return Event{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest, startsAt time.Time, endsAt time.Time) (Event, error) {
	existing, err := r.FindByID(ctx, id)
	if err != nil {
		return Event{}, err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE events
		SET organization_id = $2,
			title = $3,
			description = $4,
			format = $5,
			status = $6,
			starts_at = $7,
			ends_at = $8,
			location = $9,
			max_participants = $10,
			updated_at = now()
		WHERE id = $1
	`, id, request.OrganizationID, request.Title, request.Description, normalizeFormat(request.Format), normalizeStatus(request.Status), startsAt, endsAt, request.Location, request.MaxParticipants)
	if err != nil {
		return Event{}, err
	}
	if tag.RowsAffected() == 0 {
		return Event{}, ErrNotFound
	}
	if existing.MaxParticipants == nil || request.MaxParticipants == nil || *existing.MaxParticipants != *request.MaxParticipants {
		if err := r.RebalanceWaitlist(ctx, id); err != nil {
			return Event{}, err
		}
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func baseSelect() string {
	return `
		SELECT e.id::text, e.organization_id::text, o.name, e.title, e.description, e.format::text, e.status::text,
			e.starts_at, e.ends_at, e.location, e.max_participants, e.checkin_code, e.created_by::text, e.created_at, e.updated_at
		FROM events e
		JOIN organizations o ON o.id = e.organization_id
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(row scanner) (Event, error) {
	var item Event
	var location sql.NullString
	var maxParticipants sql.NullInt32
	var createdBy sql.NullString
	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationName,
		&item.Title,
		&item.Description,
		&item.Format,
		&item.Status,
		&item.StartsAt,
		&item.EndsAt,
		&location,
		&maxParticipants,
		&item.CheckinCode,
		&createdBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if location.Valid {
		item.Location = &location.String
	}
	if maxParticipants.Valid {
		value := int(maxParticipants.Int32)
		item.MaxParticipants = &value
	}
	if createdBy.Valid {
		item.CreatedBy = &createdBy.String
	}
	return item, err
}

func normalizeFormat(format string) string {
	switch format {
	case "online", "offline", "hybrid":
		return format
	default:
		return "offline"
	}
}

func normalizeStatus(status string) string {
	switch status {
	case "draft", "published", "completed", "cancelled":
		return status
	default:
		return "published"
	}
}
