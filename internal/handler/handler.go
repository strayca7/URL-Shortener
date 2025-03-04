package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/repository"
	"url-shortener/internal/pkg/util"

	"github.com/gin-gonic/gin"
)
func CreateShortURL(c *gin.Context) {
  var req struct { LongURL string `json:"long_url"` }
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": "invalid request"})
    return
  }
  
  // 生成短链（Base62 编码）
  shortCode := util.GenerateShortCode()
  
  // 存储到 PostgreSQL
  if err := repository.SaveURL(shortCode, req.LongURL); err != nil {
    c.JSON(500, gin.H{"error": "save failed"})
    return
  }
  
  // 缓存到 Redis（过期时间 24h）
  if err := cache.RedisCli.Set(context.Background(), shortCode, req.LongURL, 24*time.Duration(time.Hour)) ; err != nil {
	log.Fatalf("Failed to cache short URL: %v", err)
	c.JSON(500, gin.H{"error": "cache failed"})
}
  
  c.JSON(200, gin.H{"short_url": fmt.Sprintf("%s/%s", os.Getenv("DOMAIN"), shortCode)})
}

