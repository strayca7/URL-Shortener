package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"
	"url-shortener/internal/pkg/util"
	"url-shortener/internal/service"

	"github.com/gin-gonic/gin"
)

func CreateShortURL(c *gin.Context) {
	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// 生成短链（Base62 编码）
	shortCode := util.GenerateShortCode()

	// 存储到 MySQL
	if err := database.SaveURL(shortCode, req.LongURL); err != nil {
		c.JSON(500, gin.H{"error": "save failed"})
		return
	}

	// 缓存到 Redis（过期时间 24h）
	if err := cache.RedisCli.Set(context.Background(), shortCode, req.LongURL, 24*time.Duration(time.Hour)); err != nil {
		log.Fatalf("Failed to cache short URL: %v", err)
		c.JSON(500, gin.H{"error": "cache failed"})
	}

	c.JSON(200, gin.H{"short_url": fmt.Sprintf("%s/%s", os.Getenv("DOMAIN"), shortCode)})
}

func ShortenHandler(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	// 生成短码（Base62 哈希）
	shortCode := util.Base62Hash(req.URL)

	// 存储到 MySQL 和 Redis
	if err := service.SaveURL(req.URL, shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": fmt.Sprintf("%s/%s", config.Host, shortCode)})
}

func RedirectHandler(c *gin.Context) {
  shortCode := c.Param("code")
  
  // 优先从 Redis 读取
  originalURL, err := cache.GetURL(shortCode)
  if err != nil {
    // 回源查询 MySQL
    originalURL, err = database.GetURL(shortCode)
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
      return
    }
    // 回写 Redis
    if err := cache.SetURL(shortCode, originalURL); err != nil {
      log.Fatalf("Failed to cache short URL: %v", err)
    }
  }

  // 记录访问日志（异步写入数据库）
  go database.LogAccess(shortCode, c.ClientIP())
  
  c.Redirect(http.StatusFound, originalURL)
}