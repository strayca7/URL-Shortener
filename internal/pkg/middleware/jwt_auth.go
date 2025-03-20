package middleware

import (
	"net/http"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

// JwtAuth 中间件，用于 JWT 认证集成自动续期。
//
// JwtAuth middleware, used for JWT authentication and auto-renewal.
func JwtAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从请求头中获取 Access Token 和 Refresh Token
        accessToken := c.GetHeader("access_token")
        refreshToken := c.GetHeader("refresh_token")

        // 尝试解析 Access Token
        claims, err := util.ParseAccessToken(accessToken)
        if err != nil {
            // 如果 Access Token 过期，尝试使用 Refresh Token 重新签发
            if err.Error() == "access token expired" {
                refreshClaims, refreshErr := util.ParseRefreshToken(refreshToken)
                if refreshErr != nil {
                    // 如果 Refresh Token 也无效，重定向到登录页面
                    c.Redirect(http.StatusTemporaryRedirect, "/login")
                    return
                }

                // 使用 Refresh Token 的信息生成新的 Access Token 和 Refresh Token
                newAccessToken, newRefreshToken, genErr := util.GenerateTokens(refreshClaims.Subject, refreshClaims.ID)
                if genErr != nil {
                    c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate new tokens"})
                    return
                }

                // 将新的 Token 写入响应头
                c.Header("new_access_token", newAccessToken)
                c.Header("new_refresh_token", newRefreshToken)

                // 将用户信息存入上下文
                c.Set("user_id", refreshClaims.Subject)
                c.Next()
                return
            }

            // 如果是其他错误，直接返回 401 错误
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // 如果 Access Token 有效，将用户信息存入上下文
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Next()
    }
}