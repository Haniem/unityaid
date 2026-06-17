package organizations

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"
)

var validRoles = []string{"super_admin", "org_admin", "coordinator", "volunteer"}
var validStatuses = []string{"active", "inactive", "blocked"}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filters ListFilters) ([]Organization, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	return s.repository.List(ctx, filters)
}

func (s *Service) FindByID(ctx context.Context, id string) (Organization, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest) (Organization, error) {
	request = normalizeUpsert(request)
	return s.repository.Create(ctx, request)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (Organization, error) {
	request = normalizeUpsert(request)
	return s.repository.Update(ctx, id, request)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repository.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repository.SoftDelete(ctx, id)
}

func (s *Service) ListMembers(ctx context.Context, organizationID string) ([]OrganizationMember, error) {
	if _, err := s.repository.FindByID(ctx, organizationID); err != nil {
		return nil, err
	}
	return s.repository.ListMembers(ctx, organizationID)
}

func (s *Service) AddMember(ctx context.Context, organizationID string, request AddMemberRequest) (OrganizationMember, error) {
	if _, err := s.repository.FindByID(ctx, organizationID); err != nil {
		return OrganizationMember{}, err
	}
	request.Role = normalizeRole(request.Role)
	request.Status = normalizeStatus(request.Status)
	if err := validateRoleAndStatus(request.Role, request.Status); err != nil {
		return OrganizationMember{}, err
	}
	return s.repository.AddMember(ctx, organizationID, request)
}

func (s *Service) UpdateMember(ctx context.Context, organizationID string, memberID string, request UpdateMemberRequest) (OrganizationMember, error) {
	request.Role = normalizeRole(request.Role)
	request.Status = normalizeStatus(request.Status)
	if err := validateRoleAndStatus(request.Role, request.Status); err != nil {
		return OrganizationMember{}, err
	}
	return s.repository.UpdateMember(ctx, organizationID, memberID, request)
}

func (s *Service) DeleteMember(ctx context.Context, organizationID string, memberID string) error {
	return s.repository.DeleteMember(ctx, organizationID, memberID)
}

func normalizeUpsert(request UpsertRequest) UpsertRequest {
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	request.Slug = normalizeSlug(request.Slug, request.Name)
	request.ContactEmail = normalizeOptional(request.ContactEmail)
	request.LogoURL = normalizeOptional(request.LogoURL)
	request.WebsiteURL = normalizeOptional(request.WebsiteURL)
	request.Phone = normalizeOptional(request.Phone)
	request.Address = normalizeOptional(request.Address)
	return request
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

func normalizeRole(role string) string {
	return strings.TrimSpace(role)
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func validateRoleAndStatus(role string, status string) error {
	if !slices.Contains(validRoles, role) || !slices.Contains(validStatuses, status) {
		return errors.New("invalid role or status")
	}
	return nil
}

func normalizeSlug(slug string, fallback string) string {
	value := strings.ToLower(strings.TrimSpace(slug))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(fallback))
	}
	value = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "organization"
	}
	return value
}
