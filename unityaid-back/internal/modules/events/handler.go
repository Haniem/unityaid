package events

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
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить мероприятия"})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Items: items})
}

func (h *Handler) Get(c *gin.Context) {
	item, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Мероприятие не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось получить мероприятие"})
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
	if !h.requireCanManageContent(c, request.OrganizationID) {
		return
	}
	item, err := h.service.Create(c.Request.Context(), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Проверьте даты и поля мероприятия"})
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
	existing, err := h.service.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "РњРµСЂРѕРїСЂРёСЏС‚РёРµ РЅРµ РЅР°Р№РґРµРЅРѕ"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "РќРµ СѓРґР°Р»РѕСЃСЊ РїРѕР»СѓС‡РёС‚СЊ РјРµСЂРѕРїСЂРёСЏС‚РёРµ"})
		return
	}
	if !h.requireCanManageContent(c, existing.OrganizationID) {
		return
	}
	if request.OrganizationID != existing.OrganizationID {
		if !h.requireCanManageContent(c, request.OrganizationID) {
			return
		}
	}
	item, err := h.service.Update(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Мероприятие не найдено"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Проверьте даты и поля мероприятия"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Мероприятие не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Не удалось удалить мероприятие"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListApplications(c *gin.Context) {
	items, err := h.service.ListApplications(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load applications"})
		return
	}
	c.JSON(http.StatusOK, ApplicationsResponse{Items: items})
}

func (h *Handler) CreateApplication(c *gin.Context) {
	var request ApplicationRequest
	_ = c.ShouldBindJSON(&request)
	claims, _ := auth.GetClaims(c)
	item, err := h.service.CreateApplication(c.Request.Context(), c.Param("id"), request, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create application"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) UpdateApplication(c *gin.Context) {
	var request ApplicationStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	item, err := h.service.UpdateApplicationStatus(c.Request.Context(), c.Param("id"), c.Param("applicationId"), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not update application"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) DeleteApplication(c *gin.Context) {
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	if err := h.service.DeleteApplication(c.Request.Context(), c.Param("id"), c.Param("applicationId")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Application not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListAttendance(c *gin.Context) {
	items, err := h.service.ListAttendance(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load attendance"})
		return
	}
	c.JSON(http.StatusOK, AttendanceResponse{Items: items})
}

func (h *Handler) MarkAttendance(c *gin.Context) {
	var request AttendanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	item, err := h.service.MarkAttendance(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not mark attendance"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *Handler) ListShifts(c *gin.Context) {
	items, err := h.service.ListShifts(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load shifts"})
		return
	}
	c.JSON(http.StatusOK, ShiftsResponse{Items: items})
}

func (h *Handler) CreateShift(c *gin.Context) {
	var request ShiftRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	item, err := h.service.CreateShift(c.Request.Context(), c.Param("id"), request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not create shift"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) ListFeedback(c *gin.Context) {
	items, err := h.service.ListFeedback(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load feedback"})
		return
	}
	c.JSON(http.StatusOK, FeedbackResponse{Items: items})
}

func (h *Handler) CreateFeedback(c *gin.Context) {
	var request FeedbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	claims, _ := auth.GetClaims(c)
	item, err := h.service.CreateFeedback(c.Request.Context(), c.Param("id"), claims.UserID, request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Could not save feedback"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *Handler) Complete(c *gin.Context) {
	if !h.requireCanManageEvent(c, c.Param("id")) {
		return
	}
	if err := h.service.CompleteEvent(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not complete event"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) requireCanManageEvent(c *gin.Context, eventID string) bool {
	item, err := h.service.FindByID(c.Request.Context(), eventID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "РњРµСЂРѕРїСЂРёСЏС‚РёРµ РЅРµ РЅР°Р№РґРµРЅРѕ"})
			return false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "РќРµ СѓРґР°Р»РѕСЃСЊ РїСЂРѕРІРµСЂРёС‚СЊ РјРµСЂРѕРїСЂРёСЏС‚РёРµ"})
		return false
	}
	return h.requireCanManageContent(c, item.OrganizationID)
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
