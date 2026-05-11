package forms

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("form item not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) organizations(ctx context.Context) ([]PossibleValue, error) {
	rows, err := r.db.Query(ctx, "SELECT id::text, name FROM organizations WHERE deleted_at IS NULL ORDER BY name")
	return scanValues(rows, err)
}

func (r *Repository) newsCategories(ctx context.Context) ([]PossibleValue, error) {
	rows, err := r.db.Query(ctx, "SELECT id::text, name FROM news_categories ORDER BY name")
	return scanValues(rows, err)
}

func (r *Repository) skills(ctx context.Context) ([]PossibleValue, error) {
	rows, err := r.db.Query(ctx, "SELECT id::text, name FROM skills ORDER BY name")
	return scanValues(rows, err)
}

func (r *Repository) events(ctx context.Context, organizationID string) ([]PossibleValue, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id::text, title
		FROM events
		WHERE $1 = '' OR organization_id::text = $1
		ORDER BY starts_at DESC
	`, organizationID)
	return scanValues(rows, err)
}

func (r *Repository) resourceOrganizationID(ctx context.Context, entity string, id string) (string, error) {
	switch entity {
	case "organizations":
		var exists bool
		err := r.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM organizations WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return "", ErrNotFound
		}
		return id, nil
	case "news":
		return r.nullableOrganizationID(ctx, "SELECT organization_id::text FROM news WHERE id = $1", id)
	case "events":
		return r.requiredOrganizationID(ctx, "SELECT organization_id::text FROM events WHERE id = $1", id)
	case "tasks":
		return r.requiredOrganizationID(ctx, "SELECT organization_id::text FROM tasks WHERE id = $1", id)
	default:
		return "", ErrNotFound
	}
}

func (r *Repository) requiredOrganizationID(ctx context.Context, query string, id string) (string, error) {
	var organizationID string
	if err := r.db.QueryRow(ctx, query, id).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return organizationID, nil
}

func (r *Repository) nullableOrganizationID(ctx context.Context, query string, id string) (string, error) {
	var organizationID sql.NullString
	if err := r.db.QueryRow(ctx, query, id).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	if !organizationID.Valid {
		return "", nil
	}
	return organizationID.String, nil
}

func (r *Repository) organizationValues(ctx context.Context, id string) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT name, slug, description, contact_email, logo_url, website_url, phone, address
		FROM organizations
		WHERE id = $1
	`, id)
	var name, slug, description string
	var contactEmail, logoURL, websiteURL, phone, address sql.NullString
	if err := row.Scan(&name, &slug, &description, &contactEmail, &logoURL, &websiteURL, &phone, &address); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return map[string]any{
		"name":         name,
		"slug":         slug,
		"description":  description,
		"contactEmail": nullableString(contactEmail),
		"logoUrl":      nullableString(logoURL),
		"websiteUrl":   nullableString(websiteURL),
		"phone":        nullableString(phone),
		"address":      nullableString(address),
	}, nil
}

func (r *Repository) newsValues(ctx context.Context, id string) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT title, summary, cover_image_url, category_id::text, status, scheduled_at
		FROM news
		WHERE id = $1
	`, id)
	var title, summary, status string
	var coverURL, categoryID sql.NullString
	var scheduledAt sql.NullTime
	if err := row.Scan(&title, &summary, &coverURL, &categoryID, &status, &scheduledAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var scheduled any
	if scheduledAt.Valid {
		scheduled = scheduledAt.Time
	}
	return map[string]any{
		"title":         title,
		"summary":       summary,
		"coverImageUrl": nullableString(coverURL),
		"categoryId":    nullableString(categoryID),
		"status":        status,
		"scheduledAt":   scheduled,
	}, nil
}

func (r *Repository) eventValues(ctx context.Context, id string) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT organization_id::text, title, description, format::text, status::text, starts_at, ends_at, location, max_participants
		FROM events
		WHERE id = $1
	`, id)
	var organizationID, title, description, format, status string
	var startsAt, endsAt time.Time
	var location sql.NullString
	var maxParticipants sql.NullInt32
	if err := row.Scan(&organizationID, &title, &description, &format, &status, &startsAt, &endsAt, &location, &maxParticipants); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return map[string]any{
		"organizationId":  organizationID,
		"title":           title,
		"description":     description,
		"format":          format,
		"status":          status,
		"startsAt":        startsAt,
		"endsAt":          endsAt,
		"location":        nullableString(location),
		"maxParticipants": nullableInt(maxParticipants),
	}, nil
}

func (r *Repository) taskValues(ctx context.Context, id string) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT organization_id::text, event_id::text, title, description, status::text, priority, due_at
		FROM tasks
		WHERE id = $1
	`, id)
	var organizationID, title, description, status, priority string
	var eventID sql.NullString
	var dueAt sql.NullTime
	if err := row.Scan(&organizationID, &eventID, &title, &description, &status, &priority, &dueAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var due any
	if dueAt.Valid {
		due = dueAt.Time
	}
	return map[string]any{
		"organizationId": organizationID,
		"eventId":        nullableString(eventID),
		"title":          title,
		"description":    description,
		"status":         status,
		"priority":       priority,
		"dueAt":          due,
	}, nil
}

func (r *Repository) volunteerValues(ctx context.Context, userID string) (map[string]any, error) {
	row := r.db.QueryRow(ctx, `
		SELECT u.first_name, u.last_name, u.patronymic, u.avatar_url, vp.city, vp.phone, vp.bio, COALESCE(vp.status::text, 'active'), COALESCE(vp.interests, '')
		FROM users u
		LEFT JOIN volunteer_profiles vp ON vp.user_id = u.id
		WHERE u.id = $1
	`, userID)
	var firstName, lastName, status, interests string
	var patronymic, avatarURL, city, phone, bio sql.NullString
	if err := row.Scan(&firstName, &lastName, &patronymic, &avatarURL, &city, &phone, &bio, &status, &interests); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	skillIDs, err := r.volunteerSkillIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"firstName":  firstName,
		"lastName":   lastName,
		"patronymic": nullableString(patronymic),
		"avatarUrl":  nullableString(avatarURL),
		"city":       nullableString(city),
		"phone":      nullableString(phone),
		"bio":        nullableString(bio),
		"status":     status,
		"interests":  interests,
		"skillIds":   skillIDs,
	}, nil
}

func (r *Repository) volunteerSkillIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT s.id::text
		FROM volunteer_profiles vp
		JOIN volunteer_skills vs ON vs.volunteer_profile_id = vp.id
		JOIN skills s ON s.id = vs.skill_id
		WHERE vp.user_id = $1
		ORDER BY s.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func scanValues(rows pgx.Rows, err error) ([]PossibleValue, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PossibleValue{}
	for rows.Next() {
		var item PossibleValue
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func nullableString(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func nullableInt(value sql.NullInt32) any {
	if value.Valid {
		return int(value.Int32)
	}
	return nil
}
