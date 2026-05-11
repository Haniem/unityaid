package users

import "time"

type User struct {
	ID              string       `json:"id"`
	Email           string       `json:"email"`
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
	CreatedAt       time.Time    `json:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt"`
}

type SystemRole struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Membership struct {
	ID               string `json:"id"`
	OrganizationID   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	Role             string `json:"role"`
	Status           string `json:"status"`
}

type VolunteerProfile struct {
	ID            string       `json:"id"`
	UserID        string       `json:"userId"`
	Email         string       `json:"email"`
	FirstName     string       `json:"firstName"`
	LastName      string       `json:"lastName"`
	Patronymic    *string      `json:"patronymic"`
	AvatarURL     *string      `json:"avatarUrl"`
	City          *string      `json:"city"`
	Phone         *string      `json:"phone"`
	Bio           string       `json:"bio"`
	Status        string       `json:"status"`
	Interests     string       `json:"interests"`
	TotalHours    float64      `json:"totalHours"`
	Points        int          `json:"points"`
	Level         int          `json:"level"`
	Skills        []Skill      `json:"skills"`
	Organizations []Membership `json:"organizations"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type Skill struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListUsersResponse struct {
	Items []User `json:"items"`
}

type ListVolunteersResponse struct {
	Items []VolunteerProfile `json:"items"`
}

type ListSkillsResponse struct {
	Items []Skill `json:"items"`
}

type ListSystemRolesResponse struct {
	Items []SystemRole `json:"items"`
}

type UserSystemRolesResponse struct {
	Items []SystemRole `json:"items"`
}

type UserFilters struct {
	Search         string
	Role           string
	OrganizationID string
}

type VolunteerFilters struct {
	Search         string
	SkillID        string
	OrganizationID string
}

type UpdateUserRequest struct {
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	Patronymic *string `json:"patronymic"`
	AvatarURL  *string `json:"avatarUrl"`
	Locale     string  `json:"locale"`
	IsActive   *bool   `json:"isActive"`
}

type UpdateVolunteerRequest struct {
	FirstName  string   `json:"firstName" binding:"required"`
	LastName   string   `json:"lastName" binding:"required"`
	Patronymic *string  `json:"patronymic"`
	AvatarURL  *string  `json:"avatarUrl"`
	City       *string  `json:"city"`
	Phone      *string  `json:"phone"`
	Bio        string   `json:"bio"`
	Status     string   `json:"status"`
	Interests  string   `json:"interests"`
	SkillIDs   []string `json:"skillIds"`
}

type CreateSkillRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

type UpdateSystemRolesRequest struct {
	RoleIDs []string `json:"roleIds"`
}
