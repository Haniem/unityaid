package events

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) ListTemplates(ctx context.Context) ([]EventTemplate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT et.id::text, et.organization_id::text, o.name, et.name, et.title, et.description, et.format::text,
			et.location, et.max_participants, et.default_duration_minutes, et.created_at, et.updated_at
		FROM event_templates et
		JOIN organizations o ON o.id = et.organization_id
		ORDER BY et.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []EventTemplate{}
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateTemplate(ctx context.Context, request EventTemplateRequest, userID string) (EventTemplate, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO event_templates (
			organization_id, name, title, description, format, location, max_participants, default_duration_minutes, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id::text
	`, request.OrganizationID, request.Name, request.Title, request.Description, normalizeFormat(request.Format), request.Location, request.MaxParticipants, request.DefaultDurationMinutes, userID).Scan(&id)
	if err != nil {
		return EventTemplate{}, err
	}
	item, err := scanTemplate(r.db.QueryRow(ctx, `
		SELECT et.id::text, et.organization_id::text, o.name, et.name, et.title, et.description, et.format::text,
			et.location, et.max_participants, et.default_duration_minutes, et.created_at, et.updated_at
		FROM event_templates et
		JOIN organizations o ON o.id = et.organization_id
		WHERE et.id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return EventTemplate{}, ErrNotFound
	}
	return item, err
}

func scanTemplate(row scanner) (EventTemplate, error) {
	var item EventTemplate
	var location sql.NullString
	var maxParticipants sql.NullInt32
	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationName,
		&item.Name,
		&item.Title,
		&item.Description,
		&item.Format,
		&location,
		&maxParticipants,
		&item.DefaultDurationMinutes,
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
	return item, err
}
