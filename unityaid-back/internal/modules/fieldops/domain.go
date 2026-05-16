package fieldops

import "errors"

var (
	ErrNotFound = errors.New("field ops item not found")
	ErrExpired  = errors.New("qr token expired")
)

type QRToken struct {
	ID        string `json:"id"`
	EventID   string `json:"eventId"`
	Token     string `json:"token"`
	Mode      string `json:"mode"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
}

type Checkin struct {
	ID         string `json:"id"`
	EventID    string `json:"eventId"`
	UserID     string `json:"userId"`
	QRTokenID  string `json:"qrTokenId"`
	CheckinAt  string `json:"checkinAt"`
	CheckoutAt string `json:"checkoutAt"`
	Status     string `json:"status"`
	Source     string `json:"source"`
}

type CreateQRRequest struct {
	Mode      string `json:"mode"`
	ExpiresAt string `json:"expiresAt"`
}

type ScanRequest struct {
	Token string `json:"token" binding:"required"`
}
