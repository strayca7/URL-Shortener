// service 封装业务逻辑实现。
//
// service encapsulates business logic implementation.
package service

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// ShorterCodeCreater 短链创建方法，集成 Snowflake、Base62，并存储到数据库。
//
// ShorterCodeCreater creates a shorter code, integrating Snowflake and Base62,
// and stores it in the database.
func ShortCodeCreater(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	
	// email := c.GetHeader("email")

	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("error binding json:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 生成短链（Base62 编码），Snowflake 算法确保唯一性，不用去重
	shortCode, err := createShortURL()
	if err != nil {
		log.Println("error creating short URL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create short URL"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Println("error asserting userID to string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if err := database.MysqlDB.Create(&database.ShortURL{UserID: userIDStr, ShortCode: shortCode, OriginalURL: req.LongURL, ExpireAt: time.Now().Add(90 * 24 * time.Hour)}).Error; err != nil {
		log.Println("error creating short URL:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}

	// 未启用，缓存到 Redis（过期时间 24h）
	// if err := cache.SetURL(shortCode, req.LongURL); err != nil {
	// 	log.Fatalf("Failed to cache short URL: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "cache failed"})
	// }

	c.JSON(http.StatusOK, gin.H{"short_url": fmt.Sprintf("%s/%s", config.Host(), shortCode)})
}
