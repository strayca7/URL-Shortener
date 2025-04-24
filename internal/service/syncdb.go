package service

import (
	"url-shortener/internal/pkg/cache"
	"url-shortener/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// 未启用，集成 MySQL,Redis 存储URL
func RecordURL(short database.ShortURL, c *gin.Context) error {
	if err := database.CreateShortURL(short, c); err != nil {
		return err
	}
	if err := cache.SaveShortURL(short); err != nil {
		return err
	}
	return nil
}
