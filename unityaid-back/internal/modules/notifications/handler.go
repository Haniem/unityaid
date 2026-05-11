package notifications

import (
	"net/http"

	"unityaid-back/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	items, unread, err := h.service.List(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load notifications"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items, UnreadCount: unread})
}

func (h *Handler) MarkRead(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	if err := h.service.MarkRead(c.Request.Context(), claims.UserID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not mark notification as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	if err := h.service.MarkAllRead(c.Request.Context(), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not mark notifications as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
