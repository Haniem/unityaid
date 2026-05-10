package organizations

import "time"

type Organization struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	ContactEmail *string   `json:"contactEmail"`
	LogoURL      *string   `json:"logoUrl"`
	WebsiteURL   *string   `json:"websiteUrl"`
	Phone        *string   `json:"phone"`
	Address      *string   `json:"address"`
	IsDeleted    bool      `json:"isDeleted"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type OrganizationMember struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	AvatarURL *string   `json:"avatarUrl"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ListResponse struct {
	Items []Organization `json:"items"`
}

type MembersResponse struct {
	Items []OrganizationMember `json:"items"`
}

type UpsertRequest struct {
	Name         string  `json:"name" binding:"required,min=2"`
	Slug         string  `json:"slug"`
	Description  string  `json:"description"`
	ContactEmail *string `json:"contactEmail"`
	LogoURL      *string `json:"logoUrl"`
	WebsiteURL   *string `json:"websiteUrl"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
}

type ListFilters struct {
	Search         string
	IncludeDeleted bool
}

type AddMemberRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Role   string `json:"role" binding:"required"`
	Status string `json:"status"`
}

type UpdateMemberRequest struct {
	Role   string `json:"role" binding:"required"`
	Status string `json:"status" binding:"required"`
}
