package service

import (
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// 集成 MySQL,Redis 存储URL
func SaveURL(url string, shortCode string, c *gin.Context) error {
	if err := database.SaveURL(shortCode, url, c); err != nil {
		return err
	}
	if err := cache.SetURL(shortCode, url); err != nil {
		return err
	}
	return nil
}
