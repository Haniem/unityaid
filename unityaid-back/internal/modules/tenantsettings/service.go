package tenantsettings

import (
	"context"
	"regexp"
	"strings"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context) (TenantSettings, error) {
	return s.repository.Get(ctx)
}

func (s *Service) Upsert(ctx context.Context, request UpsertRequest) (TenantSettings, error) {
	request = normalizeRequest(request)
	if request.DefaultOrganizationID == nil && request.DefaultOrganizationName != "" {
		id, err := s.repository.CreateDefaultOrganization(
			ctx,
			request.DefaultOrganizationName,
			request.DefaultOrganizationSlug,
			request.Description,
			request.ContactEmail,
		)
		if err != nil {
			return TenantSettings{}, err
		}
		request.DefaultOrganizationID = &id
	}
	return s.repository.Upsert(ctx, request)
}

func normalizeRequest(request UpsertRequest) UpsertRequest {
	request.DisplayName = fallback(strings.TrimSpace(request.DisplayName), "Пульс")
	request.Description = strings.TrimSpace(request.Description)
	request.LogoURL = normalizeOptional(request.LogoURL)
	request.PrimaryColor = fallback(strings.TrimSpace(request.PrimaryColor), "#2f9f72")
	request.AccentColor = fallback(strings.TrimSpace(request.AccentColor), "#22684e")
	request.Timezone = fallback(strings.TrimSpace(request.Timezone), "Asia/Yekaterinburg")
	request.Locale = fallback(strings.TrimSpace(request.Locale), "ru")
	request.ContactEmail = normalizeOptional(request.ContactEmail)
	request.ContactPhone = normalizeOptional(request.ContactPhone)
	request.DefaultOrganizationID = normalizeOptional(request.DefaultOrganizationID)
	request.DefaultOrganizationName = strings.TrimSpace(request.DefaultOrganizationName)
	request.DefaultOrganizationSlug = normalizeSlug(request.DefaultOrganizationSlug, request.DefaultOrganizationName)
	request.PendingInvites = normalizeInvites(request.PendingInvites)
	return request
}

func normalizeInvites(items []PendingInvite) []PendingInvite {
	normalized := []PendingInvite{}
	for _, item := range items {
		item.Email = strings.ToLower(strings.TrimSpace(item.Email))
		item.Role = fallback(strings.TrimSpace(item.Role), "coordinator")
		if item.Email == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return normalized
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeSlug(slug string, fallbackName string) string {
	value := strings.ToLower(strings.TrimSpace(slug))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(fallbackName))
	}
	value = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "main"
	}
	return value
}

func fallback(value string, fallbackValue string) string {
	if value == "" {
		return fallbackValue
	}
	return value
}
