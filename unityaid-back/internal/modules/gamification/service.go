package gamification

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Profile(ctx context.Context, userID string) (Profile, error) {
	return s.repository.Profile(ctx, userID)
}

func (s *Service) Leaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	return s.repository.Leaderboard(ctx)
}

func (s *Service) ListAchievements(ctx context.Context) ([]Achievement, error) {
	return s.repository.ListAchievements(ctx)
}

func (s *Service) RecalculateAll(ctx context.Context) error {
	return s.repository.RecalculateAll(ctx)
}
