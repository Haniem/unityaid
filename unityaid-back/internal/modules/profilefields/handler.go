package profilefields

import (
	"errors"
	"log"
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

func (h *Handler) Schema(c *gin.Context) {
	org := c.Query("organizationId")
	if !h.requireAccess(c, org) {
		return
	}
	items, err := h.service.Schema(c.Request.Context(), org)
	if err != nil {
		h.internal(c)
		return
	}
	c.JSON(http.StatusOK, SchemaResponse{Items: items})
}
func (h *Handler) Values(c *gin.Context) {
	org := c.Query("organizationId")
	if !h.requireAccess(c, org) || !h.requireEditVolunteer(c, c.Param("userId")) {
		return
	}
	item, err := h.service.Values(c.Request.Context(), org, c.Param("userId"))
	if err != nil {
		h.internal(c)
		return
	}
	c.JSON(http.StatusOK, item)
}
func (h *Handler) SaveValues(c *gin.Context) {
	var request ValuesRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireAccess(c, request.OrganizationID) || !h.requireEditVolunteer(c, c.Param("userId")) {
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.SaveValues(c.Request.Context(), c.Param("userId"), claims.UserID, request)
	if err != nil {
		log.Printf("profile field values save failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Не удалось сохранить значения профиля"})
		return
	}
	c.JSON(http.StatusOK, item)
}
func (h *Handler) CreateGroup(c *gin.Context) {
	var request GroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireManage(c, request.OrganizationID) {
		return
	}
	item, err := h.service.CreateGroup(c.Request.Context(), request)
	if err != nil {
		h.internal(c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
func (h *Handler) UpdateGroup(c *gin.Context) {
	var request GroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireManage(c, request.OrganizationID) {
		return
	}
	item, err := h.service.UpdateGroup(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
func (h *Handler) DeleteGroup(c *gin.Context) {
	org := c.Query("organizationId")
	if !h.requireManage(c, org) {
		return
	}
	if err := h.service.DeleteGroup(c.Request.Context(), c.Param("id"), org); err != nil {
		h.error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) CreateField(c *gin.Context) {
	var request FieldRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireManage(c, request.OrganizationID) {
		return
	}
	item, err := h.service.CreateField(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
func (h *Handler) UpdateField(c *gin.Context) {
	var request FieldRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireManage(c, request.OrganizationID) {
		return
	}
	item, err := h.service.UpdateField(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		h.error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
func (h *Handler) DeleteField(c *gin.Context) {
	org := c.Query("organizationId")
	if !h.requireManage(c, org) {
		return
	}
	if err := h.service.DeleteField(c.Request.Context(), c.Param("id"), org); err != nil {
		h.error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) requireAccess(c *gin.Context, org string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	allowed, err := h.authorizer.CanAccessOrganization(c.Request.Context(), claims, org)
	return h.permission(c, allowed, err)
}
func (h *Handler) requireManage(c *gin.Context, org string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	allowed, err := h.authorizer.CanManageOrganization(c.Request.Context(), claims, org)
	return h.permission(c, allowed, err)
}
func (h *Handler) requireEditVolunteer(c *gin.Context, user string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		return false
	}
	allowed, err := h.authorizer.CanEditVolunteer(c.Request.Context(), claims, user)
	return h.permission(c, allowed, err)
}
func (h *Handler) permission(c *gin.Context, allowed bool, err error) bool {
	if err != nil {
		h.internal(c)
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Недостаточно прав"})
		return false
	}
	return true
}
func (h *Handler) error(c *gin.Context, err error) {
	if errors.Is(err, ErrProtected) {
		c.JSON(http.StatusConflict, gin.H{"error": "protected", "message": "Системное поле или группа защищены от удаления"})
		return
	}
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Поле не найдено"})
		return
	}
	h.internal(c)
}
func (h *Handler) internal(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось обработать профильные поля"})
}
