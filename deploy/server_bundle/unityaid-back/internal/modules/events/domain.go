package events

import "time"

type Event struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Format           string    `json:"format"`
	Status           string    `json:"status"`
	StartsAt         time.Time `json:"startsAt"`
	EndsAt           time.Time `json:"endsAt"`
	Location         *string   `json:"location"`
	MaxParticipants  *int      `json:"maxParticipants"`
	CheckinCode      string    `json:"checkinCode"`
	CreatedBy        *string   `json:"createdBy"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type ListResponse struct {
	Items []Event `json:"items"`
}

type UpsertRequest struct {
	OrganizationID  string  `json:"organizationId" binding:"required"`
	Title           string  `json:"title" binding:"required,min=3"`
	Description     string  `json:"description"`
	Format          string  `json:"format"`
	Status          string  `json:"status"`
	StartsAt        string  `json:"startsAt" binding:"required"`
	EndsAt          string  `json:"endsAt" binding:"required"`
	Location        *string `json:"location"`
	MaxParticipants *int    `json:"maxParticipants"`
}

type Application struct {
	ID              string    `json:"id"`
	EventID         string    `json:"eventId"`
	UserID          string    `json:"userId"`
	UserName        string    `json:"userName"`
	Email           string    `json:"email"`
	Status          string    `json:"status"`
	Message         string    `json:"message"`
	RejectionReason string    `json:"rejectionReason"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Attendance struct {
	ID         string     `json:"id"`
	EventID    string     `json:"eventId"`
	UserID     string     `json:"userId"`
	UserName   string     `json:"userName"`
	Email      string     `json:"email"`
	CheckInAt  *time.Time `json:"checkInAt"`
	CheckOutAt *time.Time `json:"checkOutAt"`
	Hours      float64    `json:"hours"`
}

type Shift struct {
	ID        string    `json:"id"`
	EventID   string    `json:"eventId"`
	Title     string    `json:"title"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	Capacity  *int      `json:"capacity"`
	CreatedAt time.Time `json:"createdAt"`
}

type Feedback struct {
	ID        string    `json:"id"`
	EventID   string    `json:"eventId"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type ApplicationRequest struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

type ApplicationStatusRequest struct {
	Status          string `json:"status" binding:"required"`
	RejectionReason string `json:"rejectionReason"`
}

type BulkApplicationStatusRequest struct {
	ApplicationIDs  []string `json:"applicationIds" binding:"required"`
	Status          string   `json:"status" binding:"required"`
	RejectionReason string   `json:"rejectionReason"`
}

type AttendanceRequest struct {
	UserID      string  `json:"userId" binding:"required"`
	CheckinCode string  `json:"checkinCode"`
	Hours       float64 `json:"hours"`
}

type BulkAttendanceRequest struct {
	UserIDs []string `json:"userIds" binding:"required"`
	Hours   float64  `json:"hours"`
}

type AttendanceUpdateRequest struct {
	Hours      *float64 `json:"hours"`
	CheckOutAt *string  `json:"checkOutAt"`
}

type ShiftRequest struct {
	Title    string `json:"title" binding:"required"`
	StartsAt string `json:"startsAt" binding:"required"`
	EndsAt   string `json:"endsAt" binding:"required"`
	Capacity *int   `json:"capacity"`
}

type FeedbackRequest struct {
	Rating  int    `json:"rating" binding:"required"`
	Comment string `json:"comment"`
}

type EventTemplate struct {
	ID                     string    `json:"id"`
	OrganizationID         string    `json:"organizationId"`
	OrganizationName       string    `json:"organizationName"`
	Name                   string    `json:"name"`
	Title                  string    `json:"title"`
	Description            string    `json:"description"`
	Format                 string    `json:"format"`
	Location               *string   `json:"location"`
	MaxParticipants        *int      `json:"maxParticipants"`
	DefaultDurationMinutes int       `json:"defaultDurationMinutes"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type EventTemplateRequest struct {
	OrganizationID         string  `json:"organizationId" binding:"required"`
	Name                   string  `json:"name" binding:"required"`
	Title                  string  `json:"title" binding:"required"`
	Description            string  `json:"description"`
	Format                 string  `json:"format"`
	Location               *string `json:"location"`
	MaxParticipants        *int    `json:"maxParticipants"`
	DefaultDurationMinutes int     `json:"defaultDurationMinutes"`
}

type RecurringEventRequest struct {
	EventPayload UpsertRequest `json:"event" binding:"required"`
	Frequency    string        `json:"frequency"`
	Count        int           `json:"count"`
}

type ApplicationsResponse struct {
	Items []Application `json:"items"`
}
type AttendanceResponse struct {
	Items []Attendance `json:"items"`
}
type ShiftsResponse struct {
	Items []Shift `json:"items"`
}
type FeedbackResponse struct {
	Items         []Feedback `json:"items"`
	AverageRating float64    `json:"averageRating"`
	FeedbackCount int        `json:"feedbackCount"`
}

type TemplatesResponse struct {
	Items []EventTemplate `json:"items"`
}
