package fieldops

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

func (h *Handler) CreateQRToken(c *gin.Context) {
	var request CreateQRRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.CreateQRToken(c.Request.Context(), c.Param("eventId"), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "qr_error", "message": "Could not create QR token"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Scan(c *gin.Context) {
	var request ScanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Scan(c.Request.Context(), request.Token, claims.UserID)
	if err != nil {
		status := http.StatusBadRequest
		message := "Could not scan QR token"
		if errors.Is(err, ErrNotFound) {
			status = http.StatusNotFound
			message = "QR token not found"
		}
		if errors.Is(err, ErrExpired) {
			message = "QR token expired"
		}
		c.JSON(status, gin.H{"error": "qr_error", "message": message})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
