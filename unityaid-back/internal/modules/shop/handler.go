package shop

import (
	"errors"
	"net/http"

	"unityaid-back/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *Service
	authorizer *auth.Authorizer
}

func NewHandler(service *Service, authorizer *auth.Authorizer) *Handler {
	return &Handler{service: service, authorizer: authorizer}
}

func (h *Handler) Wallet(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Wallet(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load wallet"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Transfer(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	var request TransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.Transfer(c.Request.Context(), claims.UserID, request)
	if errors.Is(err, ErrInsufficientCoins) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient_coins", "message": "Недостаточно монет"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not transfer coins"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Products(c *gin.Context) {
	items, err := h.service.Products(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	var request CreateOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.CreateOrder(c.Request.Context(), claims.UserID, request)
	if errors.Is(err, ErrInsufficientCoins) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient_coins", "message": "Недостаточно монет"})
		return
	}
	if errors.Is(err, ErrOutOfStock) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "out_of_stock", "message": "Товара недостаточно на складе"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not create order"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Orders(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	all := false
	if c.Query("scope") == "all" {
		allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not check permissions"})
			return
		}
		all = allowed
	}
	items, err := h.service.Orders(c.Request.Context(), claims.UserID, all)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load orders"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not check permissions"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Insufficient permissions"})
		return
	}
	var request UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.UpdateOrderStatus(c.Request.Context(), c.Param("id"), request.Status, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not update order"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
