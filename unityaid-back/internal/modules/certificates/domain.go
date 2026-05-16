package certificates

import "time"

type Certificate struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	UserName         string    `json:"userName"`
	Email            string    `json:"email"`
	OrganizationID   *string   `json:"organizationId"`
	OrganizationName *string   `json:"organizationName"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	TotalHours       float64   `json:"totalHours"`
	VerifyCode       string    `json:"verifyCode"`
	IssuedBy         *string   `json:"issuedBy"`
	IssuedByName     *string   `json:"issuedByName"`
	IssuedAt         time.Time `json:"issuedAt"`
	CreatedAt        time.Time `json:"createdAt"`
}

type GenerateRequest struct {
	UserID         string  `json:"userId" binding:"required"`
	OrganizationID *string `json:"organizationId"`
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
}

type ListResponse struct {
	Items []Certificate `json:"items"`
}
