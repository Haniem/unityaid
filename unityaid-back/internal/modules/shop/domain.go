package shop

import "time"

type Wallet struct {
	UserID       string            `json:"userId"`
	Balance      int               `json:"balance"`
	Transactions []CoinTransaction `json:"transactions"`
}

type CoinTransaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Amount      int       `json:"amount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Product struct {
	ID               string    `json:"id"`
	OrganizationID   *string   `json:"organizationId,omitempty"`
	OrganizationName *string   `json:"organizationName,omitempty"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Price            int       `json:"price"`
	Stock            int       `json:"stock"`
	ImageURL         *string   `json:"imageUrl,omitempty"`
	IsActive         bool      `json:"isActive"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type Order struct {
	ID               string      `json:"id"`
	UserID           string      `json:"userId"`
	UserName         string      `json:"userName"`
	OrganizationID   *string     `json:"organizationId,omitempty"`
	OrganizationName *string     `json:"organizationName,omitempty"`
	Status           string      `json:"status"`
	Total            int         `json:"total"`
	Comment          string      `json:"comment"`
	Items            []OrderItem `json:"items"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

type OrderItem struct {
	ID          string  `json:"id"`
	ProductID   *string `json:"productId,omitempty"`
	ProductName string  `json:"productName"`
	Price       int     `json:"price"`
	Quantity    int     `json:"quantity"`
}

type CreateOrderRequest struct {
	Items   []CreateOrderItem `json:"items" binding:"required"`
	Comment string            `json:"comment"`
}

type CreateOrderItem struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type TransferRequest struct {
	RecipientID string `json:"recipientId" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Comment     string `json:"comment"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
