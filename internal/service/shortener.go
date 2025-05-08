// service encapsulates business logic implementation.
package service

import (
	"net/http"
	"time"
	"url-shortener/internal/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// UserShortCodeCreater creates a shorter code, integrating Snowflake and Base62,
// and stores it in the database.
// This is a private API, so user ID is needed.
// The user ID is obtained from the JWT token in the HTTP header.
// The request body should be in JSON format, as follows:
//
//	{
//	    "long_url": "https://www.example.com"
//	}
//
// The response will be in JSON format, as follows:
//
//	{
//	    "original_url": "https://www.example.com",
//	    "short_url": "abc123"
//	}
//
// The short URL will expire in 90 days. This is default expiration time.
func UserShortCodeCreater(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// email := c.GetHeader("email")

	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Err(err).Msg("Invalid long URL request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 生成短链（Base62 编码），Snowflake 算法确保唯一性，不用去重
	shortCode, err := createShortURL()
	if err != nil {
		log.Err(err).Msg("Failed to create short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create short URL"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Warn().Msg("Error asserting userID to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if err := database.CreateUserShortURL(database.UserShortURL{UserID: userIDStr, ShortCode: shortCode, OriginalURL: req.LongURL, ExpireAt: time.Now().Add(90 * 24 * time.Hour)}, c.ClientIP()); err != nil {
		log.Warn().Err(err).Msg("Failed to create short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	// 未启用，缓存到 Redis（过期时间 24h）
	// if err := cache.SetURL(shortCode, req.LongURL); err != nil {
	// 	log.Fatalf("Failed to cache short URL: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "cache failed"})
	// }

	c.JSON(http.StatusOK, gin.H{
		"original_url": req.LongURL,
		"short_url":    shortCode,
	})
}

// PublicShortCodeCreater creates a public short code, integrating Snowflake and Base62,
// and stores it in the database.
// This is a public API, so no user ID is needed.
//
// Send JSON format as follows:
//
//	{
//	    "long_url": "https://www.example.com"
//	}
//
// Return JSON format as follows:
//
//	{
//	    "original_url": "https://www.example.com",
//	    "short_url": "abc123"
//	}
//
// The short URL will expire in 90 days.This is default expiration time.
func PublicShortCodeCreater(c *gin.Context) {
	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Err(err).Msg("Invalid long URL request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 检查 URL 是否存在
	if _, err := database.GetPublicShortURLByShortCode(req.LongURL); err == nil {
		log.Warn().Msg("URL already exists")
		c.JSON(http.StatusConflict, gin.H{"error": "URL already exists"})
		return
	}

	// 生成短链（Base62 编码），Snowflake 算法确保唯一性，不用去重
	shortCode, err := createShortURL()
	if err != nil {
		log.Err(err).Msg("Failed to create short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create short URL"})
		return
	}

	if err := database.CreatePublicShortURL(database.PublicShortURL{ShortCode: shortCode, OriginalURL: req.LongURL, ExpiresAt: time.Now().Add(90 * 24 * time.Hour)}); err != nil {
		log.Warn().Err(err).Msg("Failed to create public short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"original_url": req.LongURL,
		"short_url":    shortCode,
	})
}
