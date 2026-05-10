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
	Status           string     `json:"status"`
	AuthorID         *string    `json:"authorId"`
	AuthorName       *string    `json:"authorName"`
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
	Status         string  `json:"status"`
}
