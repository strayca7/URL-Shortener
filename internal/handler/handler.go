package handler

import (
	"errors"
	"net/http"
	"sync"

	"url-shortener/internal/pkg/database"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CreateShorterCodeHandler 短链生成接口，POST /shorten。需要在 http header 加入 Authorization 和 refresh_token，并且在 http body 中发送 JSON。
//
// CreateShorterCodeHandler API for creating short URL, POST /shorten. Requires Authorization and refresh_token in the HTTP header, and JSON in the HTTP body.
//
// 发送 JSON 格式为：/ sendsend JSON format as follows:
//
//	{
//	    "url": "https://www.example.com"
//	}
func CreateShorterCodeHandler(c *gin.Context) {
	service.ShortCodeCreater(c)
}

// RedirectHandler 短链重定向处理函数。
// 需要在 http header 加入 Authorization 和 refresh_token。
//
// 发送 http 请求，例如： http://localhost:8080/short/abc123
//
// RedirectHandler handles redirection from a short URL to the original URL.
// Requires Authorization and refresh_token in the HTTP header.
//
// send http request, for example: http://localhost:8080/short/abc123
func RedirectHandler(c *gin.Context) {
	shortCode := c.Param("code")

	originalURL, err := database.GetURL(shortCode)
	if err.Error() == "short URL has expired" {
		log.Warn().Str("shortCode", shortCode).Msg("Short URL has expired")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}
	if err != nil {
		log.Err(err).Str("shortCode", shortCode).Msg("Failed to get original URL for shortCode ")
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	clientIP := c.ClientIP()
	log.Info().Str("IP", clientIP).Msg("用户IP")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := database.LogAccess(shortCode, clientIP); err != nil {
			log.Err(err).Str("shortCode", shortCode).Msg("Failed to log access for shortCode ")
		}
	}()
	wg.Wait()

	log.Info().Str("shortCode", shortCode).Str("original URL", originalURL).Msg("Redirecting shortCode ")
	c.Redirect(http.StatusFound, originalURL)
}

// GetUserShortURLsHandler 获取用户短链接列表
func GetUserShortURLsHandler(c *gin.Context) {
	userID, exist := c.Get("user_ID")
	if !exist {
		log.Err(errors.New("userID not found")).Msg("userID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Err(errors.New("error asserting userID to string"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	shortURLs, err := database.GetUserShortURLsByUserID(userIDStr)
	if err != nil {
		log.Err(err).Str("userID", userIDStr).Msg("Failed to get short URLs for userID ")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get short URLs"})
	}

	c.JSON(http.StatusOK, shortURLs)
}
