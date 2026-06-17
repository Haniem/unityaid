package knowledge

import "time"

type Article struct {
	ID           string     `json:"id"`
	CategoryID   *string    `json:"categoryId"`
	CategoryName *string    `json:"categoryName"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Summary      string     `json:"summary"`
	ContentHTML  string     `json:"contentHtml"`
	Status       string     `json:"status"`
	AuthorID     *string    `json:"authorId"`
	AuthorName   *string    `json:"authorName"`
	PublishedAt  *time.Time `json:"publishedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type UpsertRequest struct {
	CategoryID  *string `json:"categoryId"`
	Title       string  `json:"title" binding:"required,min=3"`
	Summary     string  `json:"summary"`
	ContentHTML string  `json:"contentHtml"`
	Status      string  `json:"status"`
}

type ListFilters struct {
	Search     string
	Status     string
	CategoryID string
}

type ListResponse struct {
	Items []Article `json:"items"`
}

type CategoriesResponse struct {
	Items []Category `json:"items"`
}
