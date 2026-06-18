package invitations

import "time"

type Invitation struct {
	ID               string     `json:"id"`
	Email            *string    `json:"email"`
	Role             string     `json:"role"`
	Token            string     `json:"token"`
	Link             string     `json:"link"`
	Status           string     `json:"status"`
	OrganizationID   *string    `json:"organizationId"`
	OrganizationName *string    `json:"organizationName"`
	InvitedBy        *string    `json:"invitedBy"`
	AcceptedBy       *string    `json:"acceptedBy"`
	ExpiresAt        time.Time  `json:"expiresAt"`
	AcceptedAt       *time.Time `json:"acceptedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type CreateRequest struct {
	Email          string `json:"email"`
	Role           string `json:"role"`
	OrganizationID string `json:"organizationId"`
}

type AcceptRegistrationRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FirstName  string `json:"firstName" binding:"required"`
	LastName   string `json:"lastName" binding:"required"`
	Patronymic string `json:"patronymic"`
}

type ListResponse struct {
	Items []Invitation `json:"items"`
}
