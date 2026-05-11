package notifications

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, request CreateRequest) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, body, link, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, type, entity_type, entity_id)
		DO UPDATE SET
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			link = EXCLUDED.link,
			is_read = false,
			read_at = NULL,
			created_at = now()
	`, request.UserID, request.Type, request.Title, request.Body, request.Link, request.EntityType, request.EntityID)
	return err
}

func (r *Repository) List(ctx context.Context, userID string) ([]Notification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, type, title, body, link, entity_type, entity_id, is_read, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY is_read ASC, created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Notification{}
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`, userID).Scan(&count)
	return count, err
}

func (r *Repository) MarkRead(ctx context.Context, userID string, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET is_read = true, read_at = now()
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	return err
}

func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET is_read = true, read_at = now()
		WHERE user_id = $1 AND is_read = false
	`, userID)
	return err
}

func (r *Repository) EnsureUpcomingEventReminders(ctx context.Context, userID string) error {
	rows, err := r.db.Query(ctx, `
		SELECT e.id::text, e.title
		FROM events e
		JOIN event_applications ea ON ea.event_id = e.id
		WHERE ea.user_id = $1
			AND ea.status = 'approved'
			AND e.status = 'published'
			AND e.starts_at BETWEEN now() AND now() + interval '24 hours'
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var eventID, title string
		if err := rows.Scan(&eventID, &title); err != nil {
			return err
		}
		if err := r.Create(ctx, CreateRequest{
			UserID:     userID,
			Type:       "event_starts_soon",
			Title:      "Мероприятие скоро начнется",
			Body:       title,
			Link:       "/calendar/" + eventID,
			EntityType: "event",
			EntityID:   eventID,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(row scanner) (Notification, error) {
	var item Notification
	var readAt sql.NullTime
	err := row.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Body, &item.Link, &item.EntityType, &item.EntityID, &item.IsRead, &readAt, &item.CreatedAt)
	if readAt.Valid {
		item.ReadAt = &readAt.Time
	}
	return item, err
}
