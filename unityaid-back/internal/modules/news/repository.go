package news

import (
	"context"
	"database/sql"
	"errors"

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

func (r *Repository) List(ctx context.Context) ([]News, error) {
	rows, err := r.db.Query(ctx, baseSelect()+` ORDER BY n.created_at DESC`)
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
		INSERT INTO news (organization_id, title, slug, summary, content_html, cover_image_url, status, author_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CASE WHEN $7 = 'published' THEN now() ELSE NULL END)
		RETURNING id::text
	`, request.OrganizationID, request.Title, slug, request.Summary, request.ContentHTML, request.CoverImageURL, normalizeStatus(request.Status), authorID)

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
			status = $8,
			published_at = CASE
				WHEN $8 = 'published' AND published_at IS NULL THEN now()
				WHEN $8 <> 'published' THEN NULL
				ELSE published_at
			END,
			updated_at = now()
		WHERE id = $1
	`, id, request.OrganizationID, request.Title, slug, request.Summary, request.ContentHTML, request.CoverImageURL, normalizeStatus(request.Status))
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
			n.status,
			n.author_id::text,
			concat_ws(' ', u.last_name, u.first_name, u.patronymic),
			n.published_at,
			n.created_at,
			n.updated_at
		FROM news n
		LEFT JOIN organizations o ON o.id = n.organization_id
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
	var authorID sql.NullString
	var authorName sql.NullString
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
		&item.Status,
		&authorID,
		&authorName,
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
	if authorID.Valid {
		item.AuthorID = &authorID.String
	}
	if authorName.Valid {
		item.AuthorName = &authorName.String
	}
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}

	return item, err
}

func normalizeStatus(status string) string {
	if status == "draft" {
		return "draft"
	}
	return "published"
}
