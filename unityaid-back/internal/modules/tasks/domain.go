package tasks

import "time"

type Task struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organizationId"`
	OrganizationName string     `json:"organizationName"`
	EventID          *string    `json:"eventId"`
	EventTitle       *string    `json:"eventTitle"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	Priority         string     `json:"priority"`
	DueAt            *time.Time `json:"dueAt"`
	CreatedBy        *string    `json:"createdBy"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type ListResponse struct {
	Items []Task `json:"items"`
}

type UpsertRequest struct {
	OrganizationID string  `json:"organizationId" binding:"required"`
	EventID        *string `json:"eventId"`
	Title          string  `json:"title" binding:"required,min=3"`
	Description    string  `json:"description"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
	DueAt          *string `json:"dueAt"`
}
