package knowledge

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("knowledge article not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filters ListFilters) ([]Article, error) {
	rows, err := r.db.Query(ctx, baseSelect()+`
		WHERE ($1 = '' OR ka.status = $1)
			AND ($2 = '' OR ka.category_id::text = $2)
			AND (
				$3 = ''
				OR ka.title ILIKE '%' || $3 || '%'
				OR ka.summary ILIKE '%' || $3 || '%'
				OR ka.content_html ILIKE '%' || $3 || '%'
			)
		ORDER BY ka.published_at DESC NULLS LAST, ka.created_at DESC
	`, filters.Status, filters.CategoryID, filters.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Article{}
	for rows.Next() {
		item, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) FindByID(ctx context.Context, id string) (Article, error) {
	item, err := scanArticle(r.db.QueryRow(ctx, baseSelect()+` WHERE ka.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Article{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) Create(ctx context.Context, request UpsertRequest, slug string, authorID string) (Article, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		INSERT INTO knowledge_articles (category_id, title, slug, summary, content_html, status, author_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CASE WHEN $6 = 'published' THEN now() ELSE NULL END)
		RETURNING id::text
	`, request.CategoryID, request.Title, slug, request.Summary, request.ContentHTML, normalizeStatus(request.Status), authorID).Scan(&id)
	if err != nil {
		return Article{}, err
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id string, request UpsertRequest, slug string) (Article, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE knowledge_articles
		SET category_id = $2,
			title = $3,
			slug = $4,
			summary = $5,
			content_html = $6,
			status = $7,
			published_at = CASE
				WHEN $7 = 'published' AND published_at IS NULL THEN now()
				WHEN $7 <> 'published' THEN NULL
				ELSE published_at
			END,
			updated_at = now()
		WHERE id = $1
	`, id, request.CategoryID, request.Title, slug, request.Summary, request.ContentHTML, normalizeStatus(request.Status))
	if err != nil {
		return Article{}, err
	}
	if tag.RowsAffected() == 0 {
		return Article{}, ErrNotFound
	}
	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM knowledge_articles WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, name, slug, description FROM knowledge_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Category{}
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateCategory(ctx context.Context, name string, slug string, description string) (Category, error) {
	var item Category
	err := r.db.QueryRow(ctx, `
		INSERT INTO knowledge_categories (name, slug, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description
		RETURNING id::text, name, slug, description
	`, name, slug, description).Scan(&item.ID, &item.Name, &item.Slug, &item.Description)
	return item, err
}

func baseSelect() string {
	return `
		SELECT ka.id::text, ka.category_id::text, kc.name, ka.title, ka.slug, ka.summary, ka.content_html,
			ka.status, ka.author_id::text, concat_ws(' ', u.last_name, u.first_name),
			ka.published_at, ka.created_at, ka.updated_at
		FROM knowledge_articles ka
		LEFT JOIN knowledge_categories kc ON kc.id = ka.category_id
		LEFT JOIN users u ON u.id = ka.author_id
	`
}

type scanner interface {
	Scan(dest ...any) error
}

func scanArticle(row scanner) (Article, error) {
	var item Article
	var categoryID, categoryName, authorID, authorName sql.NullString
	var publishedAt sql.NullTime
	err := row.Scan(&item.ID, &categoryID, &categoryName, &item.Title, &item.Slug, &item.Summary, &item.ContentHTML, &item.Status, &authorID, &authorName, &publishedAt, &item.CreatedAt, &item.UpdatedAt)
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
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}
	return item, err
}

func normalizeStatus(status string) string {
	switch status {
	case "draft", "published", "archived":
		return status
	default:
		return "draft"
	}
}
