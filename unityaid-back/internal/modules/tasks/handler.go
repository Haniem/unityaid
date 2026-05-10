package tasks

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
	items, err := h.service.List(c.Request.Context(), ListFilters{Status: c.Query("status"), Priority: c.Query("priority"), AssigneeID: c.Query("assigneeId")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить задачи"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Задача не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить задачу"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Проверьте поля задачи"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Задача не найдена"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Проверьте поля задачи"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Задача не найдена"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось удалить задачу"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AddAssignment(c *gin.Context) {
	var r AssignmentRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	item, err := h.service.AddAssignment(c.Request.Context(), c.Param("id"), r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not assign task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
func (h *Handler) RemoveAssignment(c *gin.Context) {
	if err := h.service.RemoveAssignment(c.Request.Context(), c.Param("id"), c.Param("userId")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not remove assignment"})
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ListComments(c *gin.Context) {
	items, err := h.service.ListComments(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load comments"})
		return
	}
	c.JSON(http.StatusOK, CommentsResponse{Items: items})
}
func (h *Handler) AddComment(c *gin.Context) {
	var r CommentRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.AddComment(c.Request.Context(), c.Param("id"), claims.UserID, r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not add comment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
func (h *Handler) ListAttachments(c *gin.Context) {
	items, err := h.service.ListAttachments(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load attachments"})
		return
	}
	c.JSON(http.StatusOK, AttachmentsResponse{Items: items})
}
func (h *Handler) AddAttachment(c *gin.Context) {
	var r AttachmentRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.AddAttachment(c.Request.Context(), c.Param("id"), claims.UserID, r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not add attachment"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
func (h *Handler) ListStatusHistory(c *gin.Context) {
	items, err := h.service.ListStatusHistory(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load status history"})
		return
	}
	c.JSON(http.StatusOK, StatusHistoryResponse{Items: items})
}
func (h *Handler) ListTimeEntries(c *gin.Context) {
	items, err := h.service.ListTimeEntries(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load time entries"})
		return
	}
	c.JSON(http.StatusOK, TimeEntriesResponse{Items: items})
}
func (h *Handler) AddTimeEntry(c *gin.Context) {
	var r TimeEntryRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.AddTimeEntry(c.Request.Context(), c.Param("id"), claims.UserID, r)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not add time"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}
func (h *Handler) Approve(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Approve(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not approve task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}
