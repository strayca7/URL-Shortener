package router

import (
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/middleware"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_gin"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

func Router() {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		log.Info().Msg("health check")
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	public := r.Group("/public")
	{
		public.POST("/register", controller.Register)
		limiter := tollbooth.NewLimiter(5, nil) // 每秒5次请求
		public.POST("/login", tollbooth_gin.LimitHandler(limiter), controller.Login)
		public.POST("/short/new", handler.CreatePublicShortURLHandle)
		public.GET("/:code", handler.RedirectPublicCodeHandle)
		public.GET("/shortcodes", handler.GetAllPublicShortURLsHandle)
		public.DELETE("/short/:code", handler.DeletePublicShortURLHandle)
	}

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.JwtAuth())
	{
		authGroup.POST("/refresh", handler.RefreshTokenHandle)
		authGroup.POST("/short/new", handler.CreateUserShortURLHandle)
		authGroup.POST("/:code", handler.RedirectUserCodeHandle)
		authGroup.GET("/shortcodes", handler.GetUserShortURLsHandle)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
	log.Info().Msg("server started on 8080")
}
