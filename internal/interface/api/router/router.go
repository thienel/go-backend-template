package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/thienel/tlog"

	"github.com/thienel/go-backend-template/internal/interface/api/handler"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/internal/usecase/service"
)

// Setup configures all routes
func Setup(engine *gin.Engine, userService service.UserService, redisClient *redis.Client) {
	// Middleware
	engine.Use(middleware.CORS())
	engine.Use(middleware.Recovery())
	engine.Use(tlog.GinMiddleware(tlog.WithSkipPaths("/health")))
	engine.Use(middleware.RateLimiter(redisClient))

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Handlers
	userHandler := handler.NewUserHandler(userService)

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/logout", userHandler.Logout)
		}

		// User routes
		users := v1.Group("/users")
		{
			// Protected routes
			users.Use(middleware.RequireAuth())
			{
				users.GET("/me", userHandler.GetMe)
			}

			// Admin only routes
			admin := users.Group("")
			admin.Use(middleware.RequireAdmin())
			{
				admin.GET("", userHandler.List)
				admin.GET("/:id", userHandler.GetByID)
				admin.POST("", userHandler.Create)
				admin.PUT("/:id", userHandler.Update)
				admin.DELETE("/:id", userHandler.Delete)
			}
		}
	}
}
