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
func Setup(engine *gin.Engine, authService service.AuthService, userService service.UserService, redisClient *redis.Client) {
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
	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userService)

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Auth routes (protected)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.RequireAuth())
		{
			authProtected.GET("/me", authHandler.GetMe)
		}

		// User management routes (admin only)
		users := v1.Group("/users")
		users.Use(middleware.RequireAuth())
		users.Use(middleware.RequireAdmin())
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.POST("", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}
}
