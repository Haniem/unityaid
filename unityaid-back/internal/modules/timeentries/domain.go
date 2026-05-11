package timeentries

import "time"

type TimeEntry struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organizationId"`
	OrganizationName string     `json:"organizationName"`
	UserID           string     `json:"userId"`
	UserName         string     `json:"userName"`
	EventID          *string    `json:"eventId"`
	EventTitle       *string    `json:"eventTitle"`
	TaskID           *string    `json:"taskId"`
	TaskTitle        *string    `json:"taskTitle"`
	Hours            float64    `json:"hours"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	ReviewedBy       *string    `json:"reviewedBy"`
	ReviewedByName   *string    `json:"reviewedByName"`
	ReviewedAt       *time.Time `json:"reviewedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type ListFilters struct {
	Status string
	UserID string
}

type UpsertRequest struct {
	EventID     *string `json:"eventId"`
	TaskID      *string `json:"taskId"`
	Hours       float64 `json:"hours" binding:"required"`
	Description string  `json:"description"`
}

type ListResponse struct {
	Items []TimeEntry `json:"items"`
}
