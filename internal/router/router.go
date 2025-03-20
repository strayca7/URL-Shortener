package router

import (
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func Router() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	r.POST("/login", controller.Login)
	authGroup := r.Group("/api")
    authGroup.Use(middleware.JwtAuth())
    {
        authGroup.GET("/profile", func(c *gin.Context) {
            userID, _ := c.Get("user_id")
            c.JSON(200, gin.H{"user_id": userID})
        })
    }

	r.POST("/shorten", handler.ShorterCodeCreaterHandler)
	r.Run(":8080")
}
