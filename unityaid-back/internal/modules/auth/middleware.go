package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "authClaims"

func Middleware(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token", "message": "Требуется авторизация"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": "Некорректный токен"})
			return
		}

		claims, err := service.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token", "message": "Некорректный токен"})
			return
		}

		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (Claims, bool) {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return Claims{}, false
	}

	claims, ok := value.(Claims)
	return claims, ok
}
