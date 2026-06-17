package admin

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) RequireSuperAdmin(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return
	}
	allowed, err := h.authorizer.IsSuperAdmin(c.Request.Context(), claims)
	if err != nil || !allowed {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Super admin access required"})
		return
	}
	c.Next()
}

func (h *Handler) Entities(c *gin.Context) {
	c.JSON(http.StatusOK, EntitiesResponse{Items: h.service.Entities()})
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	items, err := h.service.List(c.Request.Context(), c.Param("entity"), page, limit, c.Query("search"))
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrUnknownEntity) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": "admin_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) Create(c *gin.Context) {
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.Create(c.Request.Context(), c.Param("entity"), request.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("entity"), c.Param("id"), request.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("entity"), c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin_error", "message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
