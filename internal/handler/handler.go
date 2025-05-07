package handler

import (
	"net/http"
	"sync"

	"url-shortener/internal/pkg/database"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// 短链生成接口，POST /shorten。需要在 http header 加入 Authorization 和 refresh_token，并且在 http body 中发送 JSON。
//
// API for creating short URL, POST /shorten. Requires Authorization and refresh_token in the HTTP header, and JSON in the HTTP body.
//
// 发送 JSON 格式为：/ sendsend JSON format as follows:
//
//	{
//	    "url": "https://www.example.com"
//	}
func CreateShorterCodeHandler(c *gin.Context) {
	service.ShortCodeCreater(c)
}

// 短链重定向处理函数。
// 需要在 http header 加入 Authorization 和 refresh_token。
//
// 发送 http 请求，例如： http://localhost:8080/abc123
//
// handles redirection from a short URL to the original URL.
// Requires Authorization and refresh_token in the HTTP header.
//
// send http request, for example: http://localhost:8080/abc123
func RedirectUserCodeHandler(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := database.GetOriginalURLByShortCode(shortCode)
	if err != nil {
		if err.Error() == "user short URL has expired" {
			log.Warn().Str("shortCode", shortCode).Msg("Short URL has expired")
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		log.Warn().Str("shortCode", shortCode).Msg("Failed to get original URL for shortCode ")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	clientIP := c.ClientIP()
	log.Info().Str("IP", clientIP).Msg("User IP")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.LogUserAccess(shortCode, clientIP); err != nil {
			log.Warn().Str("shortCode", shortCode).Msg("Failed to log access for shortCode ")
		}
	}()
	wg.Wait()

	log.Info().Str("shortCode", shortCode).Str("original URL", originalURL).Msg("Redirecting shortCode ")
	c.Redirect(http.StatusFound, originalURL)
}

// 公共短链重定向处理函数。
func RedirectPublicCodeHandler(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := database.GetPublicShortURLByShortCode(shortCode)
	if err != nil {
		if err.Error() == "public short URL has expired" {
			log.Warn().Str("shortCode", shortCode).Msg("Public short URL has expired")
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		log.Warn().Str("shortCode", shortCode).Msg("Failed to get original URL for shortCode ")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.LogPublicAccess(shortCode); err != nil {
			log.Warn().Str("shortCode", shortCode).Msg("Failed to log access for shortCode ")
		}
	}()
	wg.Wait()
	log.Info().Str("shortCode", shortCode).Str("original URL", originalURL).Msg("Redirecting shortCode ")
	c.Redirect(http.StatusFound, originalURL)
}

// 获取用户短链接列表
func GetUserShortURLsHandler(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		log.Warn().Msg("user ID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Warn().Msg("error asserting userID to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	shortURLs, err := database.GetUserShortURLsByUserID(userIDStr)
	if err != nil {
		log.Warn().Str("userID", userIDStr).Msg("Failed to get short URLs for userID ")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get short URLs"})
	}

	log.Info().Str("userID", userIDStr).Msg("Get short URLs for userID ")
	if len(shortURLs) == 0 {
		log.Info().Str("userID", userIDStr).Msg("No short URLs found for userID ")
		c.JSON(http.StatusOK, gin.H{"message": "No short URLs found"})
		return
	}
	c.JSON(http.StatusOK, shortURLs)
}

// 获取所有公共短链接列表
func GetAllPublicShortURLsHandler(c *gin.Context) {
	publicShortURLs, err := database.GetAllPublicShortURLs()
	if err != nil {
		log.Warn().Msg("Failed to get all public short URLs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get all public short URLs"})
		return
	}

	if len(publicShortURLs) == 0 {
		log.Info().Msg("No public short URLs found")
		c.JSON(http.StatusOK, gin.H{"message": "No public short URLs found"})
		return
	}
	c.JSON(http.StatusOK, publicShortURLs)
}

// 删除公开短链
func DeletePublicShortURLHandler(c *gin.Context) {
	shortCode := c.Param("code")

	err := database.DeletePublicShortURLByShortCode(shortCode)
	if err != nil {
		log.Warn().Str("shortCode", shortCode).Msg("Failed to delete public short URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete public short URL"})
		return
	}

	log.Info().Str("shortCode", shortCode).Msg("Deleted public short URL")
	c.JSON(http.StatusOK, gin.H{"message": "Public short URL deleted successfully"})
}