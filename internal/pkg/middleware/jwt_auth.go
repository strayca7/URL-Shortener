package middleware

import (
	"log"
	"net/http"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

// JwtAuth 中间件，用于 JWT 认证集成自动续期。
//
// JwtAuth middleware, used for JWT authentication and auto-renewal.
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.GetHeader("Authorization")
		refreshToken := c.GetHeader("refresh_token")

		claims, err := util.ParseAccessToken(accessToken)
		if err != nil {
			// 如果 Access Token 过期，尝试使用 Refresh Token 重新签发
			if err.Error() == "access token expired" {
				refreshClaims, refreshErr := util.ParseRefreshToken(refreshToken)
				if refreshErr != nil {
					// 如果 Refresh Token 无效，重定向到登录页面
					c.Redirect(http.StatusTemporaryRedirect, "/login")
					return
				}

				// 使用 Refresh Token 的信息生成新的 Access Token 和 Refresh Token
				newAccessToken, newRefreshToken, genErr := util.GenerateTokens(refreshClaims.Subject, refreshClaims.ID)
				if genErr != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate new tokens"})
					return
				}

				// 返回新的 Token 响应体， 之后需要前端自行提取 Token 并更新到请求头中
				c.JSON(http.StatusOK, gin.H{
					"access_token":  newAccessToken,
					"refresh_token": newRefreshToken,
				})

				// 将用户信息存入上下文
				c.Set("user_id", refreshClaims.Subject)
				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// 如果 Access Token 有效，将用户信息存入上下文
		log.Println("vaild Access Token")
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
