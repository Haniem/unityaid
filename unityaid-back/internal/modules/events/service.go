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
	return s.repository.UpdateApplicationStatus(ctx, eventID, applicationID, request.Status)
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

func (s *Service) CreateFeedback(ctx context.Context, eventID string, userID string, request FeedbackRequest) (Feedback, error) {
	return s.repository.CreateFeedback(ctx, eventID, userID, request)
}

func (s *Service) CompleteEvent(ctx context.Context, eventID string) error {
	return s.repository.CompleteEvent(ctx, eventID)
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
