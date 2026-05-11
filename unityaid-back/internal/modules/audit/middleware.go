package audit

import (
	"net/http"
	"strings"

	"unityaid-back/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

func Middleware(repository *Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isMutating(c.Request.Method) {
			c.Next()
			return
		}

		c.Next()

		claims, ok := auth.GetClaims(c)
		if !ok || claims.UserID == "" {
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		entry := Entry{
			UserID:     claims.UserID,
			Method:     c.Request.Method,
			Action:     action(c.Request.Method, path),
			EntityType: entityType(path),
			EntityID:   entityID(c),
			Path:       path,
			StatusCode: c.Writer.Status(),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}
		_ = repository.Record(c.Request.Context(), entry)
	}
}

func isMutating(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}

func action(method string, path string) string {
	segments := pathSegments(path)
	last := ""
	if len(segments) > 0 {
		last = segments[len(segments)-1]
	}
	switch method {
	case http.MethodPost:
		if isCommand(last) {
			return last
		}
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func entityType(path string) string {
	segments := pathSegments(path)
	if len(segments) == 0 {
		return "unknown"
	}
	return segments[0]
}

func entityID(c *gin.Context) string {
	for _, name := range []string{"id", "userId", "memberId", "applicationId", "attendanceId"} {
		if value := c.Param(name); value != "" {
			return value
		}
	}
	return ""
}

func pathSegments(path string) []string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" || part == "api" || part == "v1" || strings.HasPrefix(part, ":") {
			continue
		}
		segments = append(segments, part)
	}
	return segments
}

func isCommand(segment string) bool {
	switch segment {
	case "approve", "reject", "complete", "recalculate", "cleanup-files", "logout", "change-password", "reset-password", "verify-email", "forgot-password":
		return true
	default:
		return false
	}
}
