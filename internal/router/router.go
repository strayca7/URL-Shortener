package router

import (
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/middleware"
	"url-shortener/internal/pkg/util"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_gin"

	"github.com/gin-gonic/gin"
)

func Router() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	public := r.Group("")
	{
		public.POST("/register", controller.Register)
		limiter := tollbooth.NewLimiter(5, nil) // 每秒5次请求
		public.POST("/login", tollbooth_gin.LimitHandler(limiter), controller.Login)
	}

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.JwtAuth())
	{
		authGroup.POST("/refresh", util.RefreshTokenHandler)
		authGroup.POST("/shorten", handler.CreateShorterCodeHandler)
		authGroup.POST("/:code", handler.RedirectHandler)
	}

	r.Run(":8080")
}
