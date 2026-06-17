package notifications

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, request CreateRequest) error {
	if request.UserID == "" {
		return nil
	}
	return s.repository.Create(ctx, request)
}

func (s *Service) List(ctx context.Context, userID string) ([]Notification, int, error) {
	if err := s.repository.EnsureUpcomingEventReminders(ctx, userID); err != nil {
		return nil, 0, err
	}
	items, err := s.repository.List(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	unread, err := s.repository.CountUnread(ctx, userID)
	return items, unread, err
}

func (s *Service) MarkRead(ctx context.Context, userID string, id string) error {
	return s.repository.MarkRead(ctx, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repository.MarkAllRead(ctx, userID)
}
