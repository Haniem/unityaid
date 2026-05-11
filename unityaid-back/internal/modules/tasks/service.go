package tasks

import (
	"context"
	"time"

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
	task, err := s.repository.Update(ctx, id, request, dueAt)
	if err != nil {
		return Task{}, err
	}
	s.notifyTaskUpdated(ctx, task)
	return task, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}

func (s *Service) AddAssignment(ctx context.Context, taskID string, req AssignmentRequest) (Task, error) {
	task, err := s.repository.AddAssignment(ctx, taskID, req)
	if err != nil {
		return Task{}, err
	}
	s.notify(ctx, notifications.CreateRequest{
		UserID:     req.UserID,
		Type:       "task_assigned",
		Title:      "Назначена задача",
		Body:       task.Title,
		Link:       "/tasks/" + task.ID,
		EntityType: "task",
		EntityID:   task.ID,
	})
	return task, nil
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

func (s *Service) notifyTaskUpdated(ctx context.Context, task Task) {
	for _, assignee := range task.Assignees {
		s.notify(ctx, notifications.CreateRequest{
			UserID:     assignee.UserID,
			Type:       "task_updated",
			Title:      "Задача изменена",
			Body:       task.Title,
			Link:       "/tasks/" + task.ID,
			EntityType: "task",
			EntityID:   task.ID,
		})
	}
}

func (s *Service) notify(ctx context.Context, request notifications.CreateRequest) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Create(ctx, request)
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
