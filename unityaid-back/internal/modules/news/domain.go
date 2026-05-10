package news

import "time"

type News struct {
	ID               string     `json:"id"`
	OrganizationID   *string    `json:"organizationId"`
	OrganizationName *string    `json:"organizationName"`
	Title            string     `json:"title"`
	Slug             string     `json:"slug"`
	Summary          string     `json:"summary"`
	ContentHTML      string     `json:"contentHtml"`
	CoverImageURL    *string    `json:"coverImageUrl"`
	CategoryID       *string    `json:"categoryId"`
	CategoryName     *string    `json:"categoryName"`
	Status           string     `json:"status"`
	AuthorID         *string    `json:"authorId"`
	AuthorName       *string    `json:"authorName"`
	ScheduledAt      *time.Time `json:"scheduledAt"`
	PublishedAt      *time.Time `json:"publishedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type ListResponse struct {
	Items []News `json:"items"`
}

type UpsertRequest struct {
	OrganizationID *string `json:"organizationId"`
	Title          string  `json:"title" binding:"required,min=3"`
	Summary        string  `json:"summary"`
	ContentHTML    string  `json:"contentHtml"`
	CoverImageURL  *string `json:"coverImageUrl"`
	CategoryID     *string `json:"categoryId"`
	Status         string  `json:"status"`
	ScheduledAt    *string `json:"scheduledAt"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ListFilters struct {
	Search         string
	Status         string
	CategoryID     string
	OrganizationID string
}
type CategoriesResponse struct {
	Items []Category `json:"items"`
}
