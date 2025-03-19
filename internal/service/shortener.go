// service 封装业务逻辑实现。
//
// service encapsulates business logic implementation.
package service

import (
	"fmt"
	"log"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ShorterCodeCreater 短链创建方法，集成 Snowflake、Base62，并存储到数据库。
//
// ShorterCodeCreater creates a shorter code, integrating Snowflake and Base62, 
// and stores it in the database.
func ShortCodeCreater(c *gin.Context) {
	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 生成短链（Base62 编码），Snowflake 算法确保唯一性，不用去重
	shortCode, err := CreateShortURL()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create short URL"})
		return
	}

	// 存储到 MySQL
  if err := database.MysqlDB.Create(&database.ShortURL{ShortCode: shortCode, OriginalURL: req.LongURL}).Error; err != nil {
    c.JSON(500, gin.H{"error": "save failed"})
    return
}

	// 缓存到 Redis（过期时间 24h）
	if err := cache.SetURL(shortCode, req.LongURL); err != nil {
		log.Fatalf("Failed to cache short URL: %v", err)
		c.JSON(500, gin.H{"error": "cache failed"})
	}

	c.JSON(200, gin.H{"short_url": fmt.Sprintf("%s/%s", config.Host(), shortCode)})
}