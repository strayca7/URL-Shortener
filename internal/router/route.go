package router

import (
	"url-shortener/internal/handler"

	"github.com/gin-gonic/gin"
)

func Router() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/shorten", handler.CreateShortURL)
	r.Run(":8080")
}