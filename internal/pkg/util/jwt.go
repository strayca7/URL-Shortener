package util

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"url-shortener/internal/pkg/cache"

	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func init() {
	viper.SetConfigName("config") // 配置文件名 (不带扩展名)
	viper.SetConfigType("yaml")  // 配置文件类型
	viper.AddConfigPath("../../config")     // 配置文件路径

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

// td: 配置密钥
var SecretKey = []byte("your-256-bit-secret") // 从配置读取

// 内嵌标准声明 + 自定义业务字段
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

const (
	AccessTokenExpire  = 2 * time.Hour       // Access Token 有效期
	RefreshTokenExpire = 30 * 24 * time.Hour // Refresh Token 有效期
)

func GenerateTokens(userID string, email string) (accessToken string, refreshToken string, err error) {
	// Access Token
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
			Issuer:    "your-app-name",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(SecretKey))
	if err != nil {
		log.Println("Error generating access token:", err)
		return "", "", err
	}

	// Refresh Token（仅包含必要信息）
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpire)),
		ID:        uuid.NewString(), // 唯一标识防重放
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SecretKey))
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseAccessToken 解析并验证 AccessToken
func ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法是否匹配
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(viper.GetString("jwt.secret")), nil
		},
	)

	if err != nil {
		// 区分过期错误与其他错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("access token expired")
		}
		return nil, errors.New("invalid access token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid claims structure")
}

// ParseRefreshToken 解析并验证 RefreshToken
func ParseRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(viper.GetString("jwt.secret")), nil
		},
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("refresh token expired")
		}
		return nil, errors.New("invalid refresh token")
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		if claims.ID == "" { // 检查防重放 ID
			return nil, errors.New("missing jti in refresh token")
		}
		return claims, nil
	}

	return nil, errors.New("invalid refresh token claims")
}

// 未启用，使用 Redis 存储 Refresh Token 白名单
func StoreRefreshToken(userID string, refreshToken string) error {
	return cache.RedisCli.Set(
		context.Background(),
		fmt.Sprintf("refresh:%s", userID),
		refreshToken,
		RefreshTokenExpire,
	).Err()
}

// validateRefreshToken 验证 Refresh Token 并返回声明
func validateRefreshToken(c *gin.Context) *Claims {
	refreshToken := c.GetHeader("Authorization")
	claims, err := ParseAccessToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		c.Abort()
		return nil
	}
	return claims
}

// 刷新逻辑
func RefreshTokenHandler(c *gin.Context) {
	// 验证 Refresh Token 有效性
	claims := validateRefreshToken(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// 生成新 Token 并更新存储
	newAccessToken, newRefreshToken, _ := GenerateTokens(claims.UserID, claims.Email)
	// StoreRefreshToken(claims.UserID, newRefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}
