package forms

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

func (h *Handler) CreateForm(c *gin.Context) {
	if !h.requireCanOpenCreateForm(c, c.Param("entity")) {
		return
	}
	form, err := h.service.Get(c.Request.Context(), c.Param("entity"), "create", "")
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, form)
}

func (h *Handler) EditForm(c *gin.Context) {
	if !h.requireCanOpenEditForm(c, c.Param("entity"), c.Param("id")) {
		return
	}
	form, err := h.service.Get(c.Request.Context(), c.Param("entity"), "edit", c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, form)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Form not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load form"})
}

func (h *Handler) requireCanOpenCreateForm(c *gin.Context, entity string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}

	switch entity {
	case "organizations":
		allowed, err := h.authorizer.CanCreateOrganization(c.Request.Context(), claims)
		return h.handlePermission(c, allowed, err)
	case "news", "events", "tasks", "skills":
		allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
		return h.handlePermission(c, allowed, err)
	case "volunteers":
		return true
	default:
		return true
	}
}

func (h *Handler) requireCanOpenEditForm(c *gin.Context, entity string, id string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}

	switch entity {
	case "organizations":
		allowed, err := h.authorizer.CanManageOrganization(c.Request.Context(), claims, id)
		return h.handlePermission(c, allowed, err)
	case "news", "events", "tasks":
		organizationID, err := h.service.ResourceOrganizationID(c.Request.Context(), entity, id)
		if err != nil {
			h.handleError(c, err)
			return false
		}
		allowed, err := h.authorizer.CanManageContent(c.Request.Context(), claims, organizationID)
		return h.handlePermission(c, allowed, err)
	case "volunteers":
		allowed, err := h.authorizer.CanEditVolunteer(c.Request.Context(), claims, id)
		return h.handlePermission(c, allowed, err)
	default:
		return true
	}
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
