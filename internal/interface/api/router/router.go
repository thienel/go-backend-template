package router

import (
	"github.com/gin-gonic/gin"
	"github.com/thienel/tlog"

	"github.com/thienel/go-backend-template/internal/interface/api/handler"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
)

type routeRegister struct {
	auth   handler.AuthHandler
	user   handler.UserHandler
	policy handler.PolicyHandler
	mw     *middleware.Middleware
}

// SetupRouter configures all routes following THD-Checkin-App pattern
func SetupRouter(
	authHandler handler.AuthHandler,
	userHandler handler.UserHandler,
	policyHandler handler.PolicyHandler,
	mw *middleware.Middleware,
) *gin.Engine {

	routes := routeRegister{
		auth:   authHandler,
		user:   userHandler,
		policy: policyHandler,
		mw:     mw,
	}

	router := gin.New()
	router.Use(gin.Recovery(), mw.CORS(), tlog.GinMiddleware(tlog.WithSkipPaths("/health")))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public API
	api := router.Group("/api")
	{
		routes.registerAuthRoutes(api)
	}

	// Protected API: Authentication (JWT) + Authorization (Casbin)
	protected := api.Group("", mw.Auth(), mw.Authorize())
	{
		routes.registerUserRoutes(protected)
		routes.registerPolicyRoutes(protected)
	}

	return router
}

func (r *routeRegister) registerAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", r.auth.Login)
		auth.POST("/logout", r.auth.Logout)
	}

	// Protected auth routes
	authProtected := auth.Group("", r.mw.Auth())
	{
		authProtected.GET("/me", r.auth.GetMe)
	}
}

func (r *routeRegister) registerUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		users.GET("", r.user.List)
		users.GET("/:id", r.user.GetByID)
		users.POST("", r.user.Create)
		users.PUT("/:id", r.user.Update)
		users.DELETE("/:id", r.user.Delete)
	}
}

func (r *routeRegister) registerPolicyRoutes(rg *gin.RouterGroup) {
	policies := rg.Group("/policies")
	{
		policies.GET("", r.policy.ListPolicies)
		policies.POST("", r.policy.AddPolicy)
		policies.DELETE("", r.policy.RemovePolicy)
		policies.GET("/role/:role", r.policy.GetPoliciesForRole)
	}

	roles := rg.Group("/roles")
	{
		roles.GET("", r.policy.ListRoles)
		roles.POST("", r.policy.AddRole)
		roles.DELETE("", r.policy.RemoveRole)
	}
}
