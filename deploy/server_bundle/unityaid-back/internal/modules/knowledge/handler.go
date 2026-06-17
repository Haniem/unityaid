package knowledge

import (
	"errors"
	"net/http"

	"unityaid-back/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context(), ListFilters{
		Search:     c.Query("search"),
		Status:     c.Query("status"),
		CategoryID: c.Query("categoryId"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load knowledge articles"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load article"})
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
	item, err := h.service.Create(c.Request.Context(), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create article"})
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
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Article not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not update article"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not delete article"})
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
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.CreateCategory(c.Request.Context(), request.Name, request.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create category"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
