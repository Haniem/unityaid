package events

import "time"

type Event struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Format           string    `json:"format"`
	Status           string    `json:"status"`
	StartsAt         time.Time `json:"startsAt"`
	EndsAt           time.Time `json:"endsAt"`
	Location         *string   `json:"location"`
	MaxParticipants  *int      `json:"maxParticipants"`
	CreatedBy        *string   `json:"createdBy"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type ListResponse struct {
	Items []Event `json:"items"`
}

type UpsertRequest struct {
	OrganizationID  string  `json:"organizationId" binding:"required"`
	Title           string  `json:"title" binding:"required,min=3"`
	Description     string  `json:"description"`
	Format          string  `json:"format"`
	Status          string  `json:"status"`
	StartsAt        string  `json:"startsAt" binding:"required"`
	EndsAt          string  `json:"endsAt" binding:"required"`
	Location        *string `json:"location"`
	MaxParticipants *int    `json:"maxParticipants"`
}
