package invitations

import (
	"errors"
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
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load invitations"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Create(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return
	}
	var request CreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.Create(c.Request.Context(), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not create invitation"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Accept(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return
	}
	item, err := h.service.Accept(c.Request.Context(), c.Param("token"), claims.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Invitation not found or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not accept invitation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
