package certificates

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
	all := h.canManage(c)
	items, err := h.service.List(c.Request.Context(), claims.UserID, all)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load certificates"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Generate(c *gin.Context) {
	if !h.canManage(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Insufficient permissions"})
		return
	}
	var request GenerateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Generate(c.Request.Context(), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not generate certificate"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Download(c *gin.Context) {
	bytes, item, err := h.service.PDF(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Certificate not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not generate PDF"})
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+item.VerifyCode+`.pdf"`)
	c.Data(http.StatusOK, "application/pdf", bytes)
}

func (h *Handler) Verify(c *gin.Context) {
	item, err := h.service.FindByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Certificate not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not verify certificate"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) canManage(c *gin.Context) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		return false
	}
	allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
	return err == nil && allowed
}
