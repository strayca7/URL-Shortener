package router

import (
	"url-shortener/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/pkg/controller"
	"url-shortener/internal/pkg/middleware"
	rbacv1 "url-shortener/pkg/apis/rbac/v1"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_gin"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

func Router(rbacSys *rbacv1.RBACSystem) {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		log.Info().Msg("health check")
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	public := r.Group("/v1/public")
	{
		public.POST("/register", controller.Register)
		limiter := tollbooth.NewLimiter(5, nil) // 每秒5次请求
		public.POST("/login", tollbooth_gin.LimitHandler(limiter), controller.Login)
		public.POST("/short/new", handler.HandleCreatePublicShortURL)
		public.GET("/:code", handler.HandleRedirectPublicCode)
		public.GET("/shortcodes", handler.HandleGetAllPublicShortURLs)
		public.DELETE("/short/:code", handler.HandleDeletePublicShortURL)
	}

	authGroup := r.Group("/v1/auth")
	// If TestMod is true, then skip the JwtAuth middleware
	authGroup.Use(middleware.JwtAuth(config.TestMode))
	{
		authGroup.POST("/refresh", handler.HandleRefreshToken)
		authGroup.POST("/short/new", handler.HandleCreateUserShortURL)
		authGroup.POST("/:code", handler.HandleRedirectUserCode)
		authGroup.GET("/shortcodes", handler.HandleGetUserShortURLs)
	}

	rbacGroup := r.Group("/rbac/v1")
	{
		rbacGroup.GET("/health", rbacSys.HandleHealthCheck)
		rbacGroup.POST("/auth/check", rbacSys.HandleRBACAuthCheck)
		rbacGroup.GET("/role", rbacSys.HandleListRoles)
		rbacGroup.POST("/role", rbacSys.HandleCreateRole)
		rbacGroup.GET("/rolebinding", rbacSys.HandleListRoleBindings)
		rbacGroup.POST("/rolebinding", rbacSys.HandleCreateRoleBinding)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
	log.Info().Msg("server started on 8080")
}
