package timeentries

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("time entry not found")
var ErrInvalidRequest = errors.New("invalid time entry request")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]TimeEntry, error) {
	query := baseSelect() + ` WHERE ($1 = '' OR te.status = $1) AND ($2 = '' OR te.user_id::text = $2) ORDER BY te.created_at DESC`
	rows, err := r.db.Query(ctx, query, filters.Status, filters.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

func (r *Repository) ListForOrganizations(ctx context.Context, organizationIDs []string, filters ListFilters) ([]TimeEntry, error) {
	rows, err := r.db.Query(ctx, baseSelect()+`
		WHERE te.organization_id::text = ANY($1::text[])
			AND ($2 = '' OR te.status = $2)
			AND ($3 = '' OR te.user_id::text = $3)
		ORDER BY te.created_at DESC
	`, organizationIDs, filters.Status, filters.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

func (r *Repository) FindByID(ctx context.Context, id string) (TimeEntry, error) {
	item, err := scanEntry(r.db.QueryRow(ctx, baseSelect()+` WHERE te.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return TimeEntry{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, userID string, request UpsertRequest) (TimeEntry, error) {
	if request.Hours <= 0 || (empty(request.EventID) && empty(request.TaskID)) {
		return TimeEntry{}, ErrInvalidRequest
	}
	organizationID, err := r.ResolveOrganization(ctx, request.EventID, request.TaskID)
	if err != nil {
		return TimeEntry{}, err
	}
	var id string
	err = r.db.QueryRow(ctx, `
		INSERT INTO time_entries (organization_id, user_id, event_id, task_id, hours, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, organizationID, userID, request.EventID, request.TaskID, request.Hours, request.Description).Scan(&id)
	if err != nil {
		return TimeEntry{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest) (TimeEntry, error) {
	if request.Hours <= 0 || (empty(request.EventID) && empty(request.TaskID)) {
		return TimeEntry{}, ErrInvalidRequest
	}
	organizationID, err := r.ResolveOrganization(ctx, request.EventID, request.TaskID)
	if err != nil {
		return TimeEntry{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TimeEntry{}, err
	}
	defer tx.Rollback(ctx)

	var oldStatus string
	var oldHours float64
	var userID string
	err = tx.QueryRow(ctx, `SELECT status, hours::float8, user_id::text FROM time_entries WHERE id = $1 FOR UPDATE`, id).Scan(&oldStatus, &oldHours, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return TimeEntry{}, ErrNotFound
	}
	if err != nil {
		return TimeEntry{}, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE time_entries
		SET organization_id = $2, event_id = $3, task_id = $4, hours = $5, description = $6, updated_at = now()
		WHERE id = $1
	`, id, organizationID, request.EventID, request.TaskID, request.Hours, request.Description)
	if err != nil {
		return TimeEntry{}, err
	}
	if tag.RowsAffected() == 0 {
		return TimeEntry{}, ErrNotFound
	}
	if oldStatus == "approved" && oldHours != request.Hours {
		if _, err := tx.Exec(ctx, `
			UPDATE volunteer_profiles SET total_hours = total_hours + $2, updated_at = now() WHERE user_id = $1
		`, userID, request.Hours-oldHours); err != nil {
			return TimeEntry{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return TimeEntry{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Approve(ctx context.Context, id string, reviewerID string) (TimeEntry, error) {
	return r.review(ctx, id, reviewerID, "approved")
}

func (r *Repository) Reject(ctx context.Context, id string, reviewerID string) (TimeEntry, error) {
	return r.review(ctx, id, reviewerID, "rejected")
}

func (r *Repository) review(ctx context.Context, id string, reviewerID string, status string) (TimeEntry, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return TimeEntry{}, err
	}
	defer tx.Rollback(ctx)

	var oldStatus string
	var hours float64
	var userID string
	err = tx.QueryRow(ctx, `SELECT status, hours::float8, user_id::text FROM time_entries WHERE id = $1 FOR UPDATE`, id).Scan(&oldStatus, &hours, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return TimeEntry{}, ErrNotFound
	}
	if err != nil {
		return TimeEntry{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE time_entries
		SET status = $2, reviewed_by = $3, reviewed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, status, reviewerID); err != nil {
		return TimeEntry{}, err
	}
	if oldStatus != "approved" && status == "approved" {
		if err := r.addVolunteerHours(ctx, tx, userID, hours); err != nil {
			return TimeEntry{}, err
		}
	}
	if oldStatus == "approved" && status != "approved" {
		if err := r.addVolunteerHours(ctx, tx, userID, -hours); err != nil {
			return TimeEntry{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return TimeEntry{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) addVolunteerHours(ctx context.Context, tx pgx.Tx, userID string, hours float64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO volunteer_profiles (user_id, total_hours)
		VALUES ($1, GREATEST($2::numeric, 0))
		ON CONFLICT (user_id) DO UPDATE
		SET total_hours = GREATEST(volunteer_profiles.total_hours + $2::numeric, 0), updated_at = now()
	`, userID, hours)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `SELECT recalculate_user_gamification($1)`, userID)
	return err
}

func (r *Repository) ResolveOrganization(ctx context.Context, eventID *string, taskID *string) (string, error) {
	var organizationID string
	if empty(eventID) && empty(taskID) {
		return "", ErrInvalidRequest
	}
	if !empty(taskID) {
		err := r.db.QueryRow(ctx, `SELECT organization_id::text FROM tasks WHERE id = $1`, *taskID).Scan(&organizationID)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidRequest
		}
		if err != nil {
			return "", err
		}
		return organizationID, nil
	}
	err := r.db.QueryRow(ctx, `SELECT organization_id::text FROM events WHERE id = $1`, *eventID).Scan(&organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidRequest
	}
	return organizationID, err
}

func baseSelect() string {
	return `
		SELECT te.id::text, te.organization_id::text, o.name, te.user_id::text,
			concat_ws(' ', u.last_name, u.first_name), te.event_id::text, e.title, te.task_id::text, t.title,
			te.hours::float8, te.description, te.status, te.reviewed_by::text,
			concat_ws(' ', reviewer.last_name, reviewer.first_name), te.reviewed_at, te.created_at, te.updated_at
		FROM time_entries te
		JOIN organizations o ON o.id = te.organization_id
		JOIN users u ON u.id = te.user_id
		LEFT JOIN events e ON e.id = te.event_id
		LEFT JOIN tasks t ON t.id = te.task_id
		LEFT JOIN users reviewer ON reviewer.id = te.reviewed_by
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanEntries(rows pgx.Rows) ([]TimeEntry, error) {
	items := []TimeEntry{}
	for rows.Next() {
		item, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanEntry(row scanner) (TimeEntry, error) {
	var item TimeEntry
	var eventID, eventTitle, taskID, taskTitle, reviewedBy, reviewedByName sql.NullString
	var reviewedAt sql.NullTime
	err := row.Scan(
		&item.ID, &item.OrganizationID, &item.OrganizationName, &item.UserID, &item.UserName,
		&eventID, &eventTitle, &taskID, &taskTitle, &item.Hours, &item.Description, &item.Status,
		&reviewedBy, &reviewedByName, &reviewedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if eventID.Valid {
		item.EventID = &eventID.String
	}
	if eventTitle.Valid {
		item.EventTitle = &eventTitle.String
	}
	if taskID.Valid {
		item.TaskID = &taskID.String
	}
	if taskTitle.Valid {
		item.TaskTitle = &taskTitle.String
	}
	if reviewedBy.Valid {
		item.ReviewedBy = &reviewedBy.String
	}
	if reviewedByName.Valid && reviewedByName.String != " " {
		item.ReviewedByName = &reviewedByName.String
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	return item, err
}

func empty(value *string) bool {
	return value == nil || *value == ""
}
