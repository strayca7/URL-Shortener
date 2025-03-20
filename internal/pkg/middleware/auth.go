package middleware

import (
	"net/http"
	"time"

	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

// JwtAuth 中间件，用于 JWT 认证并自动续期。
//
// JwtAuth middleware, used for JWT authentication and auto-renewal.
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not provide Token"})
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Token"})
			return
		}

		if claims.ExpiresAt.Sub(time.Now()) < 24*time.Hour {
			newToken, _ := utils.GenerateToken(claims.UserID)
			c.Header("X-New-Token", newToken)
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
