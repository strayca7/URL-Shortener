package handler

import (
	"log"
	"net/http"

	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

// ShorterCodeCreaterHandler 短链生成接口，POST /shorten。
//
// ShorterCodeCreaterHandler API for creating short URL, POST /shorten.
func ShorterCodeCreaterHandler(c *gin.Context) {
	service.ShortCodeCreater(c)
}

func RedirectHandler(c *gin.Context) {
	shortCode := c.Param("code")

	// 优先从 Redis 读取
	originalURL, err := cache.GetURL(shortCode)
	if err != nil {
		// 回源查询 MySQL
		originalURL, err = database.GetURL(shortCode)
		if err != nil {
			log.Printf("Redis failed to get original URL: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}
		// 回写 Redis
		if err := cache.SetURL(shortCode, originalURL); err != nil {
			log.Fatalf("Redis failed to cache short URL: %v\n", err)
		}
		log.Printf("MySQL successfully got original URL: %s\n", originalURL)
		log.Printf("Redis successfully cached short URL: %s\n", originalURL)
		log.Printf("Redirecting to %s\n", originalURL)
	}

	// 记录访问日志（异步写入数据库）
	go database.LogAccess(shortCode, c.ClientIP())

	c.Redirect(http.StatusFound, originalURL)
}
