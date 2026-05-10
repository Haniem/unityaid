package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}

	response, err := h.service.Login(c.Request.Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "Неверный email или пароль"})
		case errors.Is(err, ErrInactiveUser):
			c.JSON(http.StatusForbidden, gin.H{"error": "inactive_user", "message": "Пользователь заблокирован"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось выполнить вход"})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Me(c *gin.Context) {
	claims, ok := GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Требуется авторизация"})
		return
	}

	user, err := h.service.Me(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Требуется авторизация"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
