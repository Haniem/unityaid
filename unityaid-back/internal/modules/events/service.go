package events

import (
	"context"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Event, error) {
	return s.repository.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (Event, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest, userID string) (Event, error) {
	startsAt, endsAt, err := parseRange(request)
	if err != nil {
		return Event{}, err
	}
	return s.repository.Create(ctx, request, startsAt, endsAt, userID)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (Event, error) {
	startsAt, endsAt, err := parseRange(request)
	if err != nil {
		return Event{}, err
	}
	return s.repository.Update(ctx, id, request, startsAt, endsAt)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func parseRange(request UpsertRequest) (time.Time, time.Time, error) {
	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startsAt, endsAt, nil
}
