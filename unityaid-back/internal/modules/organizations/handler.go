package organizations

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
	items, err := h.service.List(c.Request.Context(), ListFilters{
		Search:         c.Query("search"),
		IncludeDeleted: c.Query("includeDeleted") == "true",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load organizations"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load organization"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Create(c *gin.Context) {
	var request UpsertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanCreateOrganization(c) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not create organization"})
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
	if !h.requireCanManageOrganization(c, c.Param("id")) {
		return
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not update organization"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.requireCanManageOrganization(c, c.Param("id")) {
		return
	}
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not archive organization"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListMembers(c *gin.Context) {
	items, err := h.service.ListMembers(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load members"})
		return
	}
	c.JSON(http.StatusOK, MembersResponse{Items: items})
}

func (h *Handler) AddMember(c *gin.Context) {
	var request AddMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageOrganization(c, c.Param("id")) {
		return
	}

	item, err := h.service.AddMember(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User with this email was not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not add member"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) UpdateMember(c *gin.Context) {
	var request UpdateMemberRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageOrganization(c, c.Param("id")) {
		return
	}

	item, err := h.service.UpdateMember(c.Request.Context(), c.Param("id"), c.Param("memberId"), request)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Member not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not update member"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) DeleteMember(c *gin.Context) {
	if !h.requireCanManageOrganization(c, c.Param("id")) {
		return
	}
	if err := h.service.DeleteMember(c.Request.Context(), c.Param("id"), c.Param("memberId")); err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Member not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not remove member"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) requireCanCreateOrganization(c *gin.Context) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanCreateOrganization(c.Request.Context(), claims)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) requireCanManageOrganization(c *gin.Context, organizationID string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanManageOrganization(c.Request.Context(), claims, organizationID)
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
