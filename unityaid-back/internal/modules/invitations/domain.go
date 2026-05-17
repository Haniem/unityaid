package invitations

import "time"

type Invitation struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Token      string     `json:"token"`
	Link       string     `json:"link"`
	Status     string     `json:"status"`
	InvitedBy  *string    `json:"invitedBy"`
	AcceptedBy *string    `json:"acceptedBy"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	AcceptedAt *time.Time `json:"acceptedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type CreateRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"`
}

type ListResponse struct {
	Items []Invitation `json:"items"`
}
