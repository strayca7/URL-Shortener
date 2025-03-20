package util

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// td: 配置密钥
var SecretKey = []byte("your-256-bit-secret") // 从配置读取

// 内嵌标准声明 + 自定义业务字段
type Claims struct {
    UserID   string    `json:"user_id"`
    // Email string `json:"email"`
    jwt.RegisteredClaims
}

const (
    AccessTokenExpire  = 2 * time.Hour     // Access Token 有效期
    RefreshTokenExpire = 7 * 24 * time.Hour // Refresh Token 有效期
)

func GenerateToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
			Issuer:    "shorturl-service",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(SecretKey)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func RefreshToken(c *gin.Context) {
    // 验证 Refresh Token
    refreshToken := c.GetHeader("X-Refresh-Token")
    claims, err := ParseToken(refreshToken)
    if err != nil {
        c.JSON(401, gin.H{"error": "刷新凭证无效"})
        return
    }

    // 生成新 Access Token
    newAccessToken, _ := GenerateToken(claims.UserID)
    c.JSON(200, gin.H{"access_token": newAccessToken})
}