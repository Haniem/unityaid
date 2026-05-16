package admin

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Entities() []EntityConfig {
	return s.repository.Entities()
}

func (s *Service) List(ctx context.Context, code string, page int, limit int, search string) (ListResponse, error) {
	return s.repository.List(ctx, code, page, limit, search)
}

func (s *Service) Create(ctx context.Context, code string, data map[string]any) (map[string]any, error) {
	return s.repository.Create(ctx, code, data)
}

func (s *Service) Update(ctx context.Context, code string, id string, data map[string]any) (map[string]any, error) {
	return s.repository.Update(ctx, code, id, data)
}

func (s *Service) Delete(ctx context.Context, code string, id string) error {
	return s.repository.Delete(ctx, code, id)
}
