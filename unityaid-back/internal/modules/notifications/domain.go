package notifications

import "time"

type Notification struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	Type       string     `json:"type"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Link       string     `json:"link"`
	EntityType string     `json:"entityType"`
	EntityID   string     `json:"entityId"`
	IsRead     bool       `json:"isRead"`
	ReadAt     *time.Time `json:"readAt"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type CreateRequest struct {
	UserID     string
	Type       string
	Title      string
	Body       string
	Link       string
	EntityType string
	EntityID   string
}

type ListResponse struct {
	Items       []Notification `json:"items"`
	UnreadCount int            `json:"unreadCount"`
}
