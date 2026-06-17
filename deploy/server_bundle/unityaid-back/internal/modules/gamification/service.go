package gamification

import (
	"context"

	"unityaid-back/internal/modules/notifications"
)

type Service struct {
	repository *Repository
	notifier   *notifications.Service
}

func NewService(repository *Repository, notifier ...*notifications.Service) *Service {
	service := &Service{repository: repository}
	if len(notifier) > 0 {
		service.notifier = notifier[0]
	}
	return service
}

func (s *Service) Profile(ctx context.Context, userID string) (Profile, error) {
	before, _ := s.repository.UserAchievements(ctx, userID)
	profile, err := s.repository.Profile(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	s.notifyNewAchievements(ctx, userID, before, profile.Achievements)
	return profile, nil
}

func (s *Service) Leaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	return s.repository.Leaderboard(ctx)
}

func (s *Service) ListAchievements(ctx context.Context) ([]Achievement, error) {
	return s.repository.ListAchievements(ctx)
}

func (s *Service) RecalculateAll(ctx context.Context) error {
	userIDs, err := s.repository.UserIDs(ctx)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		before, _ := s.repository.UserAchievements(ctx, userID)
		if err := s.repository.RecalculateUser(ctx, userID); err != nil {
			return err
		}
		after, _ := s.repository.UserAchievements(ctx, userID)
		s.notifyNewAchievements(ctx, userID, before, after)
	}
	return nil
}

func (s *Service) notifyNewAchievements(ctx context.Context, userID string, before []UserAchievement, after []UserAchievement) {
	if s.notifier == nil {
		return
	}
	seen := map[string]bool{}
	for _, item := range before {
		seen[item.Achievement.ID] = true
	}
	for _, item := range after {
		if seen[item.Achievement.ID] {
			continue
		}
		_ = s.notifier.Create(ctx, notifications.CreateRequest{
			UserID:     userID,
			Type:       "achievement_earned",
			Title:      "Достижение получено",
			Body:       item.Achievement.Name,
			Link:       "/achievements",
			EntityType: "achievement",
			EntityID:   item.Achievement.ID,
		})
	}
}
