package tasks

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

func (s *Service) List(ctx context.Context) ([]Task, error) {
	return s.repository.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (Task, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, request UpsertRequest, userID string) (Task, error) {
	dueAt, err := parseOptionalTime(request.DueAt)
	if err != nil {
		return Task{}, err
	}
	request.EventID = normalizeOptionalID(request.EventID)
	return s.repository.Create(ctx, request, dueAt, userID)
}

func (s *Service) Update(ctx context.Context, id string, request UpsertRequest) (Task, error) {
	dueAt, err := parseOptionalTime(request.DueAt)
	if err != nil {
		return Task{}, err
	}
	request.EventID = normalizeOptionalID(request.EventID)
	return s.repository.Update(ctx, id, request, dueAt)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func parseOptionalTime(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeOptionalID(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}
