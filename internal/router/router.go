package router

import (
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/middleware"
	rbacv1 "url-shortener/pkg/rbac/v1"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_gin"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

func Router(rbac *rbacv1.RBACSystem) {
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
		public.POST("/short/new", handler.HandleCreatePublicShortURL)
		public.GET("/:code", handler.HandleRedirectPublicCode)
		public.GET("/shortcodes", handler.HandleGetAllPublicShortURLs)
		public.DELETE("/short/:code", handler.HandleDeletePublicShortURL)
	}

	authGroup := r.Group("/auth")
	authGroup.Use(middleware.JwtAuth())
	{
		authGroup.POST("/refresh", handler.HandleRefreshToken)
		authGroup.POST("/short/new", handler.HandleCreateUserShortURL)
		authGroup.POST("/:code", handler.HandleRedirectUserCode)
		authGroup.GET("/shortcodes", handler.HandleGetUserShortURLs)
	}

	rbacGroup := r.Group("/rbac/v1")
	{
		rbacGroup.POST("/auth/check", rbac.HandleRBACAuthCheck)
		rbacGroup.POST("/role", rbac.HandleCreateRole)
		rbacGroup.POST("/rolebinding", rbac.HandleCreateRoleBinding)
		rbacGroup.GET("/role", rbac.HandleListRoles)
		rbacGroup.GET("/rolebinding", rbac.HandleListRoleBindings)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
	log.Info().Msg("server started on 8080")
}
