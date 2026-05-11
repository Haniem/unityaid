package users

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

func (h *Handler) ListUsers(c *gin.Context) {
	items, err := h.service.ListUsers(c.Request.Context(), UserFilters{
		Search:         c.Query("search"),
		Role:           c.Query("role"),
		OrganizationID: c.Query("organizationId"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load users"})
		return
	}
	c.JSON(http.StatusOK, ListUsersResponse{Items: items})
}

func (h *Handler) GetUser(c *gin.Context) {
	item, err := h.service.FindUserByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var request UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageUsers(c) {
		return
	}
	item, err := h.service.UpdateUser(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) ListVolunteers(c *gin.Context) {
	items, err := h.service.ListVolunteers(c.Request.Context(), VolunteerFilters{
		Search:         c.Query("search"),
		SkillID:        c.Query("skillId"),
		OrganizationID: c.Query("organizationId"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load volunteers"})
		return
	}
	c.JSON(http.StatusOK, ListVolunteersResponse{Items: items})
}

func (h *Handler) GetVolunteer(c *gin.Context) {
	item, err := h.service.FindVolunteerByUserID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Volunteer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load volunteer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) UpdateVolunteer(c *gin.Context) {
	var request UpdateVolunteerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanEditVolunteer(c, c.Param("id")) {
		return
	}
	item, err := h.service.UpdateVolunteer(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Volunteer not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not update volunteer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) ListSkills(c *gin.Context) {
	items, err := h.service.ListSkills(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load skills"})
		return
	}
	c.JSON(http.StatusOK, ListSkillsResponse{Items: items})
}

func (h *Handler) CreateSkill(c *gin.Context) {
	var request CreateSkillRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageAnyContent(c) {
		return
	}
	item, err := h.service.CreateSkill(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create skill"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) DeleteSkill(c *gin.Context) {
	if !h.requireCanManageAnyContent(c) {
		return
	}
	if err := h.service.DeleteSkill(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrSkillNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Skill not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not delete skill"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) requireCanManageUsers(c *gin.Context) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanManageUsers(c.Request.Context(), claims)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) requireCanEditVolunteer(c *gin.Context, userID string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanEditVolunteer(c.Request.Context(), claims, userID)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) requireCanManageAnyContent(c *gin.Context) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
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
