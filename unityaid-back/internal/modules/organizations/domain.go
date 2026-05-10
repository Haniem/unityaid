package organizations

import "time"

type Organization struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	ContactEmail *string   `json:"contactEmail"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ListResponse struct {
	Items []Organization `json:"items"`
}

type UpsertRequest struct {
	Name         string  `json:"name" binding:"required,min=2"`
	Slug         string  `json:"slug"`
	Description  string  `json:"description"`
	ContactEmail *string `json:"contactEmail"`
}
