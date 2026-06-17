package knowledge

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filters ListFilters) ([]Article, error) {
	return s.repository.List(ctx, filters)
}

func (s *Service) FindByID(ctx context.Context, id string) (Article, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest, authorID string) (Article, error) {
	request.CategoryID = normalizeOptionalID(request.CategoryID)
	request.ContentHTML = sanitizeHTML(request.ContentHTML)
	return s.repository.Create(ctx, request, makeSlug(request.Title)+"-"+strconv.FormatInt(time.Now().UnixNano(), 36), authorID)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (Article, error) {
	request.CategoryID = normalizeOptionalID(request.CategoryID)
	request.ContentHTML = sanitizeHTML(request.ContentHTML)
	return s.repository.Update(ctx, id, request, makeSlug(request.Title)+"-"+shortID(id))
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repository.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, name string, description string) (Category, error) {
	name = strings.TrimSpace(name)
	return s.repository.CreateCategory(ctx, name, makeSlug(name), strings.TrimSpace(description))
}

func sanitizeHTML(value string) string {
	value = regexp.MustCompile(`(?is)<script.*?>.*?</script>`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*"[^"]*"`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*'[^']*'`).ReplaceAllString(value, "")
	value = regexp.MustCompile(`(?i)javascript:`).ReplaceAllString(value, "")
	return value
}

func makeSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "article"
	}
	return value
}

func normalizeOptionalID(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

func shortID(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}
