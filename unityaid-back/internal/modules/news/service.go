package news

import (
	"context"
	"os"
	"path/filepath"
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

func (s *Service) List(ctx context.Context, filters ListFilters) ([]News, error) {
	_ = s.repository.PublishScheduled(ctx)
	return s.repository.List(ctx, filters)
}

func (s *Service) FindByID(ctx context.Context, id string) (News, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest, authorID string, defaultOrganizationID string) (News, error) {
	if request.OrganizationID == nil && defaultOrganizationID != "" {
		request.OrganizationID = &defaultOrganizationID
	}
	request.ContentHTML = sanitizeHTML(request.ContentHTML)
	return s.repository.Create(ctx, request, makeSlug(request.Title)+"-"+strconv.FormatInt(time.Now().UnixNano(), 36), authorID)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest, defaultOrganizationID string) (News, error) {
	if request.OrganizationID == nil && defaultOrganizationID != "" {
		request.OrganizationID = &defaultOrganizationID
	}
	request.ContentHTML = sanitizeHTML(request.ContentHTML)
	return s.repository.Update(ctx, id, request, makeSlug(request.Title))
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repository.ListCategories(ctx)
}

func (s *Service) CreateCategory(ctx context.Context, name string) (Category, error) {
	return s.repository.CreateCategory(ctx, strings.TrimSpace(name), makeSlug(name))
}

func (s *Service) CleanupUnusedFiles(ctx context.Context, uploadsDir string) (int, error) {
	used, err := s.repository.UsedImageURLs(ctx)
	if err != nil {
		return 0, err
	}
	removed := 0
	err = filepath.WalkDir(uploadsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		url := "/uploads/" + filepath.ToSlash(strings.TrimPrefix(path, uploadsDir+string(os.PathSeparator)))
		if _, ok := used[url]; !ok {
			if err := os.Remove(path); err == nil {
				removed++
			}
		}
		return nil
	})
	return removed, err
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
	replacer := strings.NewReplacer(
		"а", "a", "б", "b", "в", "v", "г", "g", "д", "d", "е", "e", "ё", "e",
		"ж", "zh", "з", "z", "и", "i", "й", "y", "к", "k", "л", "l", "м", "m",
		"н", "n", "о", "o", "п", "p", "р", "r", "с", "s", "т", "t", "у", "u",
		"ф", "f", "х", "h", "ц", "c", "ч", "ch", "ш", "sh", "щ", "sch", "ъ", "",
		"ы", "y", "ь", "", "э", "e", "ю", "yu", "я", "ya",
	)
	value = replacer.Replace(value)
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "news"
	}
	return value
}
