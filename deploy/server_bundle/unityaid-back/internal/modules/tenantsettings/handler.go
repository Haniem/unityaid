package tenantsettings

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load tenant settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.Upsert(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not save tenant settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
