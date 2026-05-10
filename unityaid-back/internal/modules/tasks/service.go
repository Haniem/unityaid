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

func (s *Service) List(ctx context.Context, filters ListFilters) ([]Task, error) {
	return s.repository.List(ctx, filters)
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

func (s *Service) AddAssignment(ctx context.Context, taskID string, req AssignmentRequest) (Task, error) {
	return s.repository.AddAssignment(ctx, taskID, req)
}
func (s *Service) RemoveAssignment(ctx context.Context, taskID, userID string) error {
	return s.repository.RemoveAssignment(ctx, taskID, userID)
}
func (s *Service) AddComment(ctx context.Context, taskID, userID string, req CommentRequest) (TaskComment, error) {
	return s.repository.AddComment(ctx, taskID, userID, req)
}
func (s *Service) ListComments(ctx context.Context, taskID string) ([]TaskComment, error) {
	return s.repository.ListComments(ctx, taskID)
}
func (s *Service) AddAttachment(ctx context.Context, taskID, userID string, req AttachmentRequest) (TaskAttachment, error) {
	return s.repository.AddAttachment(ctx, taskID, userID, req)
}
func (s *Service) ListAttachments(ctx context.Context, taskID string) ([]TaskAttachment, error) {
	return s.repository.ListAttachments(ctx, taskID)
}
func (s *Service) ListStatusHistory(ctx context.Context, taskID string) ([]TaskStatusHistory, error) {
	return s.repository.ListStatusHistory(ctx, taskID)
}
func (s *Service) AddTimeEntry(ctx context.Context, taskID, userID string, req TimeEntryRequest) (TaskTimeEntry, error) {
	return s.repository.AddTimeEntry(ctx, taskID, userID, req)
}
func (s *Service) ListTimeEntries(ctx context.Context, taskID string) ([]TaskTimeEntry, error) {
	return s.repository.ListTimeEntries(ctx, taskID)
}
func (s *Service) Approve(ctx context.Context, taskID, userID string) (Task, error) {
	return s.repository.Approve(ctx, taskID, userID)
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
