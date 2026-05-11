package analytics

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Overview(ctx context.Context, filters Filters) (OverviewReport, error) {
	return s.repository.Overview(ctx, filters)
}

func (s *Service) Volunteers(ctx context.Context, filters Filters) (VolunteersReport, error) {
	return s.repository.Volunteers(ctx, filters)
}

func (s *Service) Events(ctx context.Context, filters Filters) (EventsReport, error) {
	return s.repository.Events(ctx, filters)
}

func (s *Service) Tasks(ctx context.Context, filters Filters) (TasksReport, error) {
	return s.repository.Tasks(ctx, filters)
}

func (s *Service) Gamification(ctx context.Context, filters Filters) (GamificationReport, error) {
	return s.repository.Gamification(ctx, filters)
}

func (s *Service) Audit(ctx context.Context, filters Filters) (AuditReport, error) {
	return s.repository.Audit(ctx, filters)
}
