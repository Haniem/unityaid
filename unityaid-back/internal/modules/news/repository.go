package news

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("news not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]News, error) {
	rows, err := r.db.Query(ctx, baseSelect()+`
		WHERE ($1 = '' OR n.title ILIKE '%' || $1 || '%' OR n.summary ILIKE '%' || $1 || '%' OR n.content_html ILIKE '%' || $1 || '%')
			AND ($2 = '' OR n.status = $2)
			AND ($3 = '' OR n.category_id::text = $3)
			AND ($4 = '' OR n.organization_id::text = $4)
		ORDER BY COALESCE(n.scheduled_at, n.published_at, n.created_at) DESC
	`, filters.Search, filters.Status, filters.CategoryID, filters.OrganizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []News{}
	for rows.Next() {
		item, err := scanNews(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (News, error) {
	row := r.db.QueryRow(ctx, baseSelect()+` WHERE n.id = $1`, id)
	item, err := scanNews(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return News{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest, slug string, authorID string) (News, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO news (organization_id, title, slug, summary, content_html, cover_image_url, category_id, status, author_id, scheduled_at, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CASE WHEN $8 = 'published' AND $10 IS NULL THEN now() ELSE NULL END)
		RETURNING id::text
	`, request.OrganizationID, request.Title, slug, request.Summary, request.ContentHTML, request.CoverImageURL, request.CategoryID, normalizeStatus(request.Status), authorID, parseSQLTime(request.ScheduledAt))

	var id string
	if err := row.Scan(&id); err != nil {
		return News{}, err
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest, slug string) (News, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE news
		SET organization_id = $2,
			title = $3,
			slug = $4,
			summary = $5,
			content_html = $6,
			cover_image_url = $7,
			category_id = $8,
			status = $9,
			scheduled_at = $10,
			published_at = CASE
				WHEN $9 = 'published' AND $10 IS NULL AND published_at IS NULL THEN now()
				WHEN $9 <> 'published' THEN NULL
				ELSE published_at
			END,
			updated_at = now()
		WHERE id = $1
	`, id, request.OrganizationID, request.Title, slug, request.Summary, request.ContentHTML, request.CoverImageURL, request.CategoryID, normalizeStatus(request.Status), parseSQLTime(request.ScheduledAt))
	if err != nil {
		return News{}, err
	}
	if tag.RowsAffected() == 0 {
		return News{}, ErrNotFound
	}

	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM news WHERE id = $1`, id)
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
		SELECT
			n.id::text,
			n.organization_id::text,
			o.name,
			n.title,
			n.slug,
			n.summary,
			n.content_html,
			n.cover_image_url,
			n.category_id::text,
			nc.name,
			n.status,
			n.author_id::text,
			concat_ws(' ', u.last_name, u.first_name, u.patronymic),
			n.scheduled_at,
			n.published_at,
			n.created_at,
			n.updated_at
		FROM news n
		LEFT JOIN organizations o ON o.id = n.organization_id
		LEFT JOIN news_categories nc ON nc.id = n.category_id
		LEFT JOIN users u ON u.id = n.author_id
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNews(row scanner) (News, error) {
	var item News
	var organizationID sql.NullString
	var organizationName sql.NullString
	var coverImageURL sql.NullString
	var categoryID sql.NullString
	var categoryName sql.NullString
	var authorID sql.NullString
	var authorName sql.NullString
	var scheduledAt sql.NullTime
	var publishedAt sql.NullTime

	err := row.Scan(
		&item.ID,
		&organizationID,
		&organizationName,
		&item.Title,
		&item.Slug,
		&item.Summary,
		&item.ContentHTML,
		&coverImageURL,
		&categoryID,
		&categoryName,
		&item.Status,
		&authorID,
		&authorName,
		&scheduledAt,
		&publishedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if organizationID.Valid {
		item.OrganizationID = &organizationID.String
	}
	if organizationName.Valid {
		item.OrganizationName = &organizationName.String
	}
	if coverImageURL.Valid {
		item.CoverImageURL = &coverImageURL.String
	}
	if categoryID.Valid {
		item.CategoryID = &categoryID.String
	}
	if categoryName.Valid {
		item.CategoryName = &categoryName.String
	}
	if authorID.Valid {
		item.AuthorID = &authorID.String
	}
	if authorName.Valid {
		item.AuthorName = &authorName.String
	}
	if scheduledAt.Valid {
		item.ScheduledAt = &scheduledAt.Time
	}
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}

	return item, err
}

func normalizeStatus(status string) string {
	switch status {
	case "draft", "scheduled":
		return status
	case "published":
		return "published"
	default:
		return "published"
	}
}

func parseSQLTime(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil
	}
	return parsed
}

func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, name, slug FROM news_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Category{}
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, name string, slug string) (Category, error) {
	var item Category
	err := r.db.QueryRow(ctx, `
		INSERT INTO news_categories (name, slug) VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
		RETURNING id::text, name, slug
	`, name, slug).Scan(&item.ID, &item.Name, &item.Slug)
	return item, err
}

func (r *Repository) PublishScheduled(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `UPDATE news SET status = 'published', published_at = now(), updated_at = now() WHERE status = 'scheduled' AND scheduled_at <= now()`)
	return err
}

func (r *Repository) UsedImageURLs(ctx context.Context) (map[string]struct{}, error) {
	rows, err := r.db.Query(ctx, `SELECT cover_image_url, content_html FROM news`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	used := map[string]struct{}{}
	for rows.Next() {
		var cover, html sql.NullString
		if err := rows.Scan(&cover, &html); err != nil {
			return nil, err
		}
		if cover.Valid {
			used[cover.String] = struct{}{}
		}
		if html.Valid {
			for _, match := range regexp.MustCompile(`/uploads/[^"'<> ]+`).FindAllString(html.String, -1) {
				used[match] = struct{}{}
			}
		}
	}
	return used, rows.Err()
}
