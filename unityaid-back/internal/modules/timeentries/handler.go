package timeentries

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

func (h *Handler) List(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	items, err := h.service.List(c.Request.Context(), claims, ListFilters{Status: c.Query("status"), UserID: c.Query("userId")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load time entries"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Create(c *gin.Context) {
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	organizationID, err := h.service.ResolveOrganization(c.Request.Context(), request)
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	if !h.requireCanAccessOrganization(c, organizationID) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), claims.UserID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create time entry"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Update(c *gin.Context) {
	existing, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	claims, _ := auth.GetClaims(c)
	if existing.UserID != claims.UserID && !h.requireCanManageEntry(c, existing) {
		return
	}
	if existing.Status != "pending" && existing.UserID == claims.UserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Only pending entries can be edited by volunteer"})
		return
	}
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	targetOrganizationID, err := h.service.ResolveOrganization(c.Request.Context(), request)
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	if existing.UserID == claims.UserID {
		if !h.requireCanAccessOrganization(c, targetOrganizationID) {
			return
		}
	} else if !h.requireCanManageOrganization(c, targetOrganizationID) {
		return
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Approve(c *gin.Context) {
	if !h.requireCanReview(c, c.Param("id")) {
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Approve(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Reject(c *gin.Context) {
	if !h.requireCanReview(c, c.Param("id")) {
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Reject(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		h.handleEntryError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) requireCanReview(c *gin.Context, id string) bool {
	item, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		h.handleEntryError(c, err)
		return false
	}
	return h.requireCanManageEntry(c, item)
}

func (h *Handler) requireCanManageEntry(c *gin.Context, item TimeEntry) bool {
	return h.requireCanManageOrganization(c, item.OrganizationID)
}

func (h *Handler) requireCanManageOrganization(c *gin.Context, organizationID string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanManageContent(c.Request.Context(), claims, organizationID)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) requireCanAccessOrganization(c *gin.Context, organizationID string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanAccessOrganization(c.Request.Context(), claims, organizationID)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) handlePermission(c *gin.Context, allowed bool, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not check permissions"})
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Insufficient permissions"})
		return false
	}
	return true
}

func (h *Handler) handleEntryError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Time entry not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not process time entry"})
}
