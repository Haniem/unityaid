package gamification

import (
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

func (h *Handler) Me(c *gin.Context) {
	claims, _ := auth.GetClaims(c)
	item, err := h.service.Profile(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load gamification profile"})
		return
	}
	c.JSON(http.StatusOK, MeResponse{Item: item})
}

func (h *Handler) UserProfile(c *gin.Context) {
	item, err := h.service.Profile(c.Request.Context(), c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load gamification profile"})
		return
	}
	c.JSON(http.StatusOK, MeResponse{Item: item})
}

func (h *Handler) Leaderboard(c *gin.Context) {
	items, err := h.service.Leaderboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load leaderboard"})
		return
	}
	c.JSON(http.StatusOK, LeaderboardResponse{Items: items})
}

func (h *Handler) ListAchievements(c *gin.Context) {
	items, err := h.service.ListAchievements(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not load achievements"})
		return
	}
	c.JSON(http.StatusOK, AchievementsResponse{Items: items})
}

func (h *Handler) Recalculate(c *gin.Context) {
	claims, ok := auth.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Authorization is required"})
		return
	}
	allowed, err := h.authorizer.CanManageAnyContent(c.Request.Context(), claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not check permissions"})
		return
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "Insufficient permissions"})
		return
	}
	if err := h.service.RecalculateAll(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Could not recalculate achievements"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
