package tenantsettings

import "time"

type PendingInvite struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TenantSettings struct {
	ID                      string          `json:"id"`
	DisplayName             string          `json:"displayName"`
	Description             string          `json:"description"`
	LogoURL                 *string         `json:"logoUrl"`
	PrimaryColor            string          `json:"primaryColor"`
	AccentColor             string          `json:"accentColor"`
	Timezone                string          `json:"timezone"`
	Locale                  string          `json:"locale"`
	ContactEmail            *string         `json:"contactEmail"`
	ContactPhone            *string         `json:"contactPhone"`
	DefaultOrganizationID   *string         `json:"defaultOrganizationId"`
	DefaultOrganizationName string          `json:"defaultOrganizationName"`
	DefaultOrganizationSlug string          `json:"defaultOrganizationSlug"`
	OnboardingCompleted     bool            `json:"onboardingCompleted"`
	PendingInvites          []PendingInvite `json:"pendingInvites"`
	CreatedAt               time.Time       `json:"createdAt"`
	UpdatedAt               time.Time       `json:"updatedAt"`
}

type UpsertRequest struct {
	DisplayName             string          `json:"displayName" binding:"required,min=2"`
	Description             string          `json:"description"`
	LogoURL                 *string         `json:"logoUrl"`
	PrimaryColor            string          `json:"primaryColor"`
	AccentColor             string          `json:"accentColor"`
	Timezone                string          `json:"timezone"`
	Locale                  string          `json:"locale"`
	ContactEmail            *string         `json:"contactEmail"`
	ContactPhone            *string         `json:"contactPhone"`
	DefaultOrganizationID   *string         `json:"defaultOrganizationId"`
	DefaultOrganizationName string          `json:"defaultOrganizationName"`
	DefaultOrganizationSlug string          `json:"defaultOrganizationSlug"`
	OnboardingCompleted     bool            `json:"onboardingCompleted"`
	PendingInvites          []PendingInvite `json:"pendingInvites"`
}
