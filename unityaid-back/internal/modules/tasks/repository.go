package tasks

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("task not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]Task, error) {
	rows, err := r.db.Query(ctx, baseSelect()+`
		WHERE ($1 = '' OR t.status::text = $1)
			AND ($2 = '' OR t.priority = $2)
			AND ($3 = '' OR EXISTS (SELECT 1 FROM task_assignments ta WHERE ta.task_id = t.id AND ta.user_id::text = $3))
		ORDER BY t.created_at DESC
	`, filters.Status, filters.Priority, filters.AssigneeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Task{}
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range items {
		_ = r.loadAssignees(ctx, &items[i])
	}
	return items, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (Task, error) {
	item, err := scanTask(r.db.QueryRow(ctx, baseSelect()+` WHERE t.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err == nil {
		_ = r.loadAssignees(ctx, &item)
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest, dueAt *time.Time, userID string) (Task, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO tasks (organization_id, event_id, title, description, status, priority, due_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text
	`, request.OrganizationID, request.EventID, request.Title, request.Description, normalizeStatus(request.Status), normalizePriority(request.Priority), dueAt, userID).Scan(&id)
	if err != nil {
		return Task{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest, dueAt *time.Time) (Task, error) {
	current, _ := r.FindByID(ctx, id)
	tag, err := r.db.Exec(ctx, `
		UPDATE tasks
		SET organization_id = $2,
			event_id = $3,
			title = $4,
			description = $5,
			status = $6,
			priority = $7,
			due_at = $8,
			updated_at = now()
		WHERE id = $1
	`, id, request.OrganizationID, request.EventID, request.Title, request.Description, normalizeStatus(request.Status), normalizePriority(request.Priority), dueAt)
	if err != nil {
		return Task{}, err
	}
	if tag.RowsAffected() == 0 {
		return Task{}, ErrNotFound
	}
	if current.Status != "" && current.Status != normalizeStatus(request.Status) {
		_, _ = r.db.Exec(ctx, `INSERT INTO task_status_history (task_id, from_status, to_status) VALUES ($1, $2, $3)`, id, current.Status, normalizeStatus(request.Status))
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
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
		SELECT t.id::text, t.organization_id::text, o.name, t.event_id::text, e.title,
			t.title, t.description, t.status::text, t.priority, t.due_at, t.created_by::text,
			t.completion_confirmed_by::text, t.completion_confirmed_at, t.created_at, t.updated_at
		FROM tasks t
		JOIN organizations o ON o.id = t.organization_id
		LEFT JOIN events e ON e.id = t.event_id
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(row scanner) (Task, error) {
	var item Task
	var eventID sql.NullString
	var eventTitle sql.NullString
	var dueAt sql.NullTime
	var createdBy sql.NullString
	var confirmedBy sql.NullString
	var confirmedAt sql.NullTime
	err := row.Scan(
		&item.ID,
		&item.OrganizationID,
		&item.OrganizationName,
		&eventID,
		&eventTitle,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Priority,
		&dueAt,
		&createdBy,
		&confirmedBy,
		&confirmedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if eventID.Valid {
		item.EventID = &eventID.String
	}
	if eventTitle.Valid {
		item.EventTitle = &eventTitle.String
	}
	if dueAt.Valid {
		item.DueAt = &dueAt.Time
	}
	if createdBy.Valid {
		item.CreatedBy = &createdBy.String
	}
	if confirmedBy.Valid {
		item.ConfirmedBy = &confirmedBy.String
	}
	if confirmedAt.Valid {
		item.ConfirmedAt = &confirmedAt.Time
	}
	return item, err
}

func (r *Repository) loadAssignees(ctx context.Context, task *Task) error {
	rows, err := r.db.Query(ctx, `
		SELECT u.id::text, concat_ws(' ', u.last_name, u.first_name), u.email, ta.role
		FROM task_assignments ta JOIN users u ON u.id = ta.user_id WHERE ta.task_id = $1 ORDER BY ta.role, u.last_name
	`, task.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	task.Assignees = []TaskAssignee{}
	for rows.Next() {
		var item TaskAssignee
		if err := rows.Scan(&item.UserID, &item.Name, &item.Email, &item.Role); err != nil {
			return err
		}
		task.Assignees = append(task.Assignees, item)
	}
	return rows.Err()
}

func normalizeStatus(status string) string {
	switch status {
	case "created", "assigned", "in_progress", "review", "completed", "cancelled":
		return status
	default:
		return "created"
	}
}

func normalizePriority(priority string) string {
	switch priority {
	case "low", "medium", "high":
		return priority
	default:
		return "medium"
	}
}
