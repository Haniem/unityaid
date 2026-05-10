package auth

import "time"

type User struct {
	ID             string     `json:"id"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	FirstName      string     `json:"firstName"`
	LastName       string     `json:"lastName"`
	Patronymic     *string    `json:"patronymic"`
	AvatarURL      *string    `json:"avatarUrl"`
	Locale         string     `json:"locale"`
	IsActive       bool       `json:"isActive"`
	LastLoginAt    *time.Time `json:"lastLoginAt"`
	Organizations  []Membership `json:"organizations"`
	PrimaryRole    string     `json:"primaryRole"`
	OrganizationID *string    `json:"organizationId"`
}

type Membership struct {
	OrganizationID   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	Role             string `json:"role"`
	Status           string `json:"status"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
	User        User   `json:"user"`
}

type Claims struct {
	UserID         string
	Email          string
	Role           string
	OrganizationID string
}
