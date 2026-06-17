package events

import (
	"context"
	"errors"
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

func (s *Service) CreateRecurring(ctx context.Context, request RecurringEventRequest, userID string) ([]Event, error) {
	startsAt, endsAt, err := parseRange(request.EventPayload)
	if err != nil {
		return nil, err
	}
	count := request.Count
	if count < 1 {
		count = 1
	}
	if count > 24 {
		count = 24
	}
	items := make([]Event, 0, count)
	for index := 0; index < count; index++ {
		payload := request.EventPayload
		nextStart := addRecurrence(startsAt, request.Frequency, index)
		nextEnd := addRecurrence(endsAt, request.Frequency, index)
		item, err := s.repository.Create(ctx, payload, nextStart, nextEnd, userID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
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

func (s *Service) ListApplications(ctx context.Context, eventID string) ([]Application, error) {
	return s.repository.ListApplications(ctx, eventID)
}

func (s *Service) CreateApplication(ctx context.Context, eventID string, request ApplicationRequest, currentUserID string) (Application, error) {
	userID := request.UserID
	if userID == "" {
		userID = currentUserID
	}
	return s.repository.CreateApplication(ctx, eventID, userID, request.Message)
}

func (s *Service) UpdateApplicationStatus(ctx context.Context, eventID string, applicationID string, request ApplicationStatusRequest) (Application, error) {
	application, err := s.repository.UpdateApplicationStatus(ctx, eventID, applicationID, request.Status, request.RejectionReason)
	if err != nil {
		return Application{}, err
	}
	if application.Status == "approved" || application.Status == "rejected" {
		event, findErr := s.repository.FindByID(ctx, eventID)
		title := "Заявка обновлена"
		if application.Status == "approved" {
			title = "Заявка подтверждена"
		}
		if application.Status == "rejected" {
			title = "Заявка отклонена"
		}
		body := application.Status
		if findErr == nil {
			body = event.Title
		}
		s.notify(ctx, notifications.CreateRequest{
			UserID:     application.UserID,
			Type:       "application_" + application.Status,
			Title:      title,
			Body:       body,
			Link:       "/calendar/" + eventID,
			EntityType: "event_application",
			EntityID:   application.ID,
		})
	}
	return application, nil
}

func (s *Service) BulkUpdateApplications(ctx context.Context, eventID string, request BulkApplicationStatusRequest) ([]Application, error) {
	items, err := s.repository.BulkUpdateApplications(ctx, eventID, request)
	if err != nil {
		return nil, err
	}
	for _, application := range items {
		if application.Status == "approved" || application.Status == "rejected" {
			event, findErr := s.repository.FindByID(ctx, eventID)
			body := application.Status
			if findErr == nil {
				body = event.Title
			}
			s.notify(ctx, notifications.CreateRequest{
				UserID:     application.UserID,
				Type:       "application_" + application.Status,
				Title:      "Заявка обновлена",
				Body:       body,
				Link:       "/calendar/" + eventID,
				EntityType: "event_application",
				EntityID:   application.ID,
			})
		}
	}
	return items, nil
}

func (s *Service) DeleteApplication(ctx context.Context, eventID string, applicationID string) error {
	return s.repository.DeleteApplication(ctx, eventID, applicationID)
}

func (s *Service) DeleteOwnApplication(ctx context.Context, eventID string, applicationID string, userID string) error {
	return s.repository.DeleteApplicationForUser(ctx, eventID, applicationID, userID)
}

func (s *Service) ListAttendance(ctx context.Context, eventID string) ([]Attendance, error) {
	return s.repository.ListAttendance(ctx, eventID)
}

func (s *Service) MarkAttendance(ctx context.Context, eventID string, request AttendanceRequest) (Attendance, error) {
	return s.repository.MarkAttendance(ctx, eventID, request)
}

func (s *Service) BulkMarkAttendance(ctx context.Context, eventID string, request BulkAttendanceRequest) ([]Attendance, error) {
	return s.repository.BulkMarkAttendance(ctx, eventID, request)
}

func (s *Service) UpdateAttendance(ctx context.Context, eventID string, attendanceID string, request AttendanceUpdateRequest) (Attendance, error) {
	var checkOutAt *time.Time
	if request.CheckOutAt != nil && *request.CheckOutAt != "" {
		parsed, err := time.Parse(time.RFC3339, *request.CheckOutAt)
		if err != nil {
			return Attendance{}, err
		}
		checkOutAt = &parsed
	}
	return s.repository.UpdateAttendance(ctx, eventID, attendanceID, request.Hours, checkOutAt)
}

func (s *Service) ListShifts(ctx context.Context, eventID string) ([]Shift, error) {
	return s.repository.ListShifts(ctx, eventID)
}

func (s *Service) CreateShift(ctx context.Context, eventID string, request ShiftRequest) (Shift, error) {
	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		return Shift{}, err
	}
	endsAt, err := time.Parse(time.RFC3339, request.EndsAt)
	if err != nil {
		return Shift{}, err
	}
	return s.repository.CreateShift(ctx, eventID, request, startsAt, endsAt)
}

func (s *Service) ListFeedback(ctx context.Context, eventID string) ([]Feedback, error) {
	return s.repository.ListFeedback(ctx, eventID)
}

func (s *Service) FeedbackSummary(ctx context.Context, eventID string) (float64, int, error) {
	return s.repository.FeedbackSummary(ctx, eventID)
}

func (s *Service) CreateFeedback(ctx context.Context, eventID string, userID string, request FeedbackRequest) (Feedback, error) {
	if request.Rating < 1 || request.Rating > 5 {
		return Feedback{}, errors.New("rating must be between 1 and 5")
	}
	return s.repository.CreateFeedback(ctx, eventID, userID, request)
}

func (s *Service) CompleteEvent(ctx context.Context, eventID string) error {
	return s.repository.CompleteEvent(ctx, eventID)
}

func (s *Service) ListTemplates(ctx context.Context) ([]EventTemplate, error) {
	return s.repository.ListTemplates(ctx)
}

func (s *Service) CreateTemplate(ctx context.Context, request EventTemplateRequest, userID string) (EventTemplate, error) {
	if request.DefaultDurationMinutes <= 0 {
		request.DefaultDurationMinutes = 120
	}
	return s.repository.CreateTemplate(ctx, request, userID)
}

func (s *Service) notify(ctx context.Context, request notifications.CreateRequest) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Create(ctx, request)
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

func addRecurrence(value time.Time, frequency string, index int) time.Time {
	switch frequency {
	case "monthly":
		return value.AddDate(0, index, 0)
	case "daily":
		return value.AddDate(0, 0, index)
	default:
		return value.AddDate(0, 0, index*7)
	}
}
