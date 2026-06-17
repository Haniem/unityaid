package tasks

import "time"

type Task struct {
	ID               string         `json:"id"`
	OrganizationID   string         `json:"organizationId"`
	OrganizationName string         `json:"organizationName"`
	EventID          *string        `json:"eventId"`
	EventTitle       *string        `json:"eventTitle"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Status           string         `json:"status"`
	Priority         string         `json:"priority"`
	DueAt            *time.Time     `json:"dueAt"`
	CreatedBy        *string        `json:"createdBy"`
	Assignees        []TaskAssignee `json:"assignees"`
	ConfirmedBy      *string        `json:"confirmedBy"`
	ConfirmedAt      *time.Time     `json:"confirmedAt"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

type ListResponse struct {
	Items []Task `json:"items"`
}

type UpsertRequest struct {
	OrganizationID string  `json:"organizationId" binding:"required"`
	EventID        *string `json:"eventId"`
	Title          string  `json:"title" binding:"required,min=3"`
	Description    string  `json:"description"`
	Status         string  `json:"status"`
	Priority       string  `json:"priority"`
	DueAt          *string `json:"dueAt"`
}

type TaskAssignee struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type TaskComment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
type TaskAttachment struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	FileURL   string    `json:"fileUrl"`
	CreatedAt time.Time `json:"createdAt"`
}
type TaskStatusHistory struct {
	ID         string    `json:"id"`
	FromStatus *string   `json:"fromStatus"`
	ToStatus   string    `json:"toStatus"`
	UserName   *string   `json:"userName"`
	CreatedAt  time.Time `json:"createdAt"`
}
type TaskTimeEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Hours     float64   `json:"hours"`
	Note      string    `json:"note"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListFilters struct {
	Status     string
	Priority   string
	AssigneeID string
}
type AssignmentRequest struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role"`
}
type CommentRequest struct {
	Content string `json:"content" binding:"required"`
}
type AttachmentRequest struct {
	FileName string `json:"fileName" binding:"required"`
	FileURL  string `json:"fileUrl" binding:"required"`
}
type TimeEntryRequest struct {
	Hours float64 `json:"hours" binding:"required"`
	Note  string  `json:"note"`
}

type CommentsResponse struct {
	Items []TaskComment `json:"items"`
}
type AttachmentsResponse struct {
	Items []TaskAttachment `json:"items"`
}
type StatusHistoryResponse struct {
	Items []TaskStatusHistory `json:"items"`
}
type TimeEntriesResponse struct {
	Items []TaskTimeEntry `json:"items"`
}
