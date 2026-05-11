package users

import (
	"context"
	"strings"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListUsers(ctx context.Context, filters UserFilters) ([]User, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.Role = strings.TrimSpace(filters.Role)
	filters.OrganizationID = strings.TrimSpace(filters.OrganizationID)
	return s.repository.ListUsers(ctx, filters)
}

func (s *Service) FindUserByID(ctx context.Context, id string) (User, error) {
	return s.repository.FindUserByID(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, id string, request UpdateUserRequest) (User, error) {
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Patronymic = normalizeOptional(request.Patronymic)
	request.AvatarURL = normalizeOptional(request.AvatarURL)
	request.Locale = strings.TrimSpace(request.Locale)
	if request.Locale == "" {
		request.Locale = "ru"
	}
	return s.repository.UpdateUser(ctx, id, request)
}

func (s *Service) ListVolunteers(ctx context.Context, filters VolunteerFilters) ([]VolunteerProfile, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.SkillID = strings.TrimSpace(filters.SkillID)
	filters.OrganizationID = strings.TrimSpace(filters.OrganizationID)
	return s.repository.ListVolunteers(ctx, filters)
}

func (s *Service) FindVolunteerByUserID(ctx context.Context, userID string) (VolunteerProfile, error) {
	return s.repository.FindVolunteerByUserID(ctx, userID)
}

func (s *Service) UpdateVolunteer(ctx context.Context, userID string, request UpdateVolunteerRequest) (VolunteerProfile, error) {
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Patronymic = normalizeOptional(request.Patronymic)
	request.AvatarURL = normalizeOptional(request.AvatarURL)
	request.City = normalizeOptional(request.City)
	request.Phone = normalizeOptional(request.Phone)
	request.Bio = strings.TrimSpace(request.Bio)
	request.SkillIDs = uniqueTrimmed(request.SkillIDs)
	return s.repository.UpdateVolunteer(ctx, userID, request)
}

func (s *Service) ListSkills(ctx context.Context) ([]Skill, error) {
	return s.repository.ListSkills(ctx)
}

func (s *Service) CreateSkill(ctx context.Context, request CreateSkillRequest) (Skill, error) {
	request.Name = strings.TrimSpace(request.Name)
	return s.repository.CreateSkill(ctx, request)
}

func (s *Service) DeleteSkill(ctx context.Context, id string) error {
	return s.repository.DeleteSkill(ctx, id)
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

func uniqueTrimmed(values []string) []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
