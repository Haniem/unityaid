package organizations

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

func (s *Service) List(ctx context.Context) ([]Organization, error) {
	return s.repository.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (Organization, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest) (Organization, error) {
	request.Slug = normalizeSlug(request.Slug, request.Name)
	return s.repository.Create(ctx, request)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (Organization, error) {
	request.Slug = normalizeSlug(request.Slug, request.Name)
	return s.repository.Update(ctx, id, request)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func normalizeSlug(slug string, fallback string) string {
	value := strings.ToLower(strings.TrimSpace(slug))
	if value == "" {
		value = strings.ToLower(strings.TrimSpace(fallback))
	}
	value = regexp.MustCompile(`[^a-z0-9а-яё]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "organization"
	}
	return value
}
