package auth

import "time"

type User struct {
	ID              string       `json:"id"`
	Email           string       `json:"email"`
	PasswordHash    string       `json:"-"`
	FirstName       string       `json:"firstName"`
	LastName        string       `json:"lastName"`
	Patronymic      *string      `json:"patronymic"`
	AvatarURL       *string      `json:"avatarUrl"`
	Locale          string       `json:"locale"`
	IsEmailVerified bool         `json:"isEmailVerified"`
	IsActive        bool         `json:"isActive"`
	LastLoginAt     *time.Time   `json:"lastLoginAt"`
	Organizations   []Membership `json:"organizations"`
	SystemRoles     []SystemRole `json:"systemRoles"`
	PrimaryRole     string       `json:"primaryRole"`
	OrganizationID  *string      `json:"organizationId"`
}

type SystemRole struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"`
	User         User   `json:"user"`
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

type AuthResponse = LoginResponse

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type DevTokenResponse struct {
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
}

type Claims struct {
	UserID         string
	Email          string
	Role           string
	OrganizationID string
	JTI            string
	ExpiresAt      time.Time
}
