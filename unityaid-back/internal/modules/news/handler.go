package news

import (
	"errors"
	"net/http"

	"unityaid-back/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    *Service
	uploadsDir string
	authorizer *auth.Authorizer
}

func NewHandler(service *Service, uploadsDir string, authorizer *auth.Authorizer) *Handler {
	return &Handler{service: service, uploadsDir: uploadsDir, authorizer: authorizer}
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), ListFilters{Search: c.Query("search"), Status: c.Query("status"), CategoryID: c.Query("categoryId"), OrganizationID: c.Query("organizationId")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить новости"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Новость не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить новость"})
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

	claims, _ := auth.GetClaims(c)
	targetOrganizationID := claims.OrganizationID
	if request.OrganizationID != nil {
		targetOrganizationID = *request.OrganizationID
	}
	if !h.requireCanManageContent(c, targetOrganizationID) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), request, claims.UserID, claims.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось создать новость"})
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

	claims, _ := auth.GetClaims(c)
	existing, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "РќРѕРІРѕСЃС‚СЊ РЅРµ РЅР°Р№РґРµРЅР°"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "РќРµ СѓРґР°Р»РѕСЃСЊ РїРѕР»СѓС‡РёС‚СЊ РЅРѕРІРѕСЃС‚СЊ"})
		return
	}
	if !h.requireCanManageNews(c, claims, existing.OrganizationID) {
		return
	}
	if request.OrganizationID != nil && !sameStringPointer(existing.OrganizationID, request.OrganizationID) {
		if !h.requireCanManageContent(c, *request.OrganizationID) {
			return
		}
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), request, claims.OrganizationID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Новость не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось обновить новость"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	existing, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "РќРѕРІРѕСЃС‚СЊ РЅРµ РЅР°Р№РґРµРЅР°"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "РќРµ СѓРґР°Р»РѕСЃСЊ РїРѕР»СѓС‡РёС‚СЊ РЅРѕРІРѕСЃС‚СЊ"})
		return
	}
	claims, _ := auth.GetClaims(c)
	if !h.requireCanManageNews(c, claims, existing.OrganizationID) {
		return
	}
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Новость не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось удалить новость"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListCategories(c *gin.Context) {
	items, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load categories"})
		return
	}
	c.JSON(http.StatusOK, CategoriesResponse{Items: items})
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var request struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageAnyContent(c) {
		return
	}
	item, err := h.service.CreateCategory(c.Request.Context(), request.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create category"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) CleanupFiles(c *gin.Context) {
	if !h.requireCanManageAnyContent(c) {
		return
	}
	removed, err := h.service.CleanupUnusedFiles(c.Request.Context(), h.uploadsDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not cleanup files"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"removed": removed})
}

func (h *Handler) requireCanManageNews(c *gin.Context, claims auth.Claims, organizationID *string) bool {
	if organizationID == nil {
		return h.requireCanManageContent(c, "")
	}
	allowed, err := h.authorizer.CanManageContent(c.Request.Context(), claims, *organizationID)
	return h.handlePermission(c, allowed, err)
}

func (h *Handler) requireCanManageContent(c *gin.Context, organizationID string) bool {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return false
	}
	allowed, err := h.authorizer.CanManageContent(c.Request.Context(), claims, organizationID)
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

func sameStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
