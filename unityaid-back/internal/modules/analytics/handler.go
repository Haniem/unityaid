package analytics

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

func (h *Handler) Overview(c *gin.Context) {
	item, err := h.service.Overview(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load overview report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Volunteers(c *gin.Context) {
	item, err := h.service.Volunteers(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load volunteers report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Events(c *gin.Context) {
	item, err := h.service.Events(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load events report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Tasks(c *gin.Context) {
	item, err := h.service.Tasks(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load tasks report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Gamification(c *gin.Context) {
	item, err := h.service.Gamification(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load gamification report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Audit(c *gin.Context) {
	item, err := h.service.Audit(c.Request.Context(), filters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load audit report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func filters(c *gin.Context) Filters {
	return Filters{From: c.Query("from"), To: c.Query("to")}
}
