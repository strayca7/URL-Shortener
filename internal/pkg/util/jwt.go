package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

const (
	AccessTokenExpire  = 2 * time.Hour
	RefreshTokenExpire = 30 * 24 * time.Hour
)

func GenerateTokens(userID string, email string) (accessToken string, refreshToken string, err error) {
	// Access Token
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
			Issuer:    "shortener",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Err(err).Msg("Error generating access token")
		return "", "", err
	}

	// Refresh Token（仅包含必要信息）
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpire)),
		ID:        uuid.NewString(),
		Subject:   userID,
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(viper.GetString("jwt_secret")))
	if err != nil {
		log.Err(err).Msg("Error generating refresh token")
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseAccessToken 解析并验证 AccessToken
func ParseAccessToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法是否匹配
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(viper.GetString("jwt_secret")), nil
		},
	)

	if err != nil {
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
			return []byte(viper.GetString("jwt_secret")), nil
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
	return cache.Rdb.Set(
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

// RefreshToken check Refresh Token's validity. If valid, generate new Access Token and Refresh Token
func RefreshToken(c *gin.Context) {
	// 验证 Refresh Token 有效性
	claims := validateRefreshToken(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	// 生成新 Token 并更新存储
	newAccessToken, newRefreshToken, err := GenerateTokens(claims.UserID, claims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}
	// StoreRefreshToken(claims.UserID, newRefreshToken)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// LoginResponse 用于后端解析登录响应
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	} `json:"user"`
}
