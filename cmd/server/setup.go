package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thienel/tlog"
	"go.uber.org/zap"

	"github.com/thienel/go-backend-template/internal/infra/authorization"
	"github.com/thienel/go-backend-template/internal/infra/database"
	"github.com/thienel/go-backend-template/internal/infra/persistence"
	"github.com/thienel/go-backend-template/internal/interface/api/handler"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/internal/interface/api/router"
	"github.com/thienel/go-backend-template/internal/usecase/service/serviceimpl"
	"github.com/thienel/go-backend-template/pkg/config"
)

// setupDependencies wires up all layers and returns the configured router
func setupDependencies(cfg *config.Config) *gin.Engine {
	// Repositories
	client := database.GetClient()
	userRepo := persistence.NewUserRepository(client)

	// Casbin Enforcer
	enforcer, err := authorization.NewEnforcer(cfg.Casbin.ModelPath, database.GetDB())
	if err != nil {
		tlog.Fatal("Failed to initialize Casbin enforcer", zap.Error(err))
	}

	// Services
	jwtService := serviceimpl.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiryMinutes,
		cfg.JWT.RefreshExpiryHours,
	)
	authService := serviceimpl.NewAuthService(userRepo, jwtService)
	userService := serviceimpl.NewUserService(userRepo)
	authzService := serviceimpl.NewAuthorizationService(enforcer)

	// Middleware
	origins := strings.Join(cfg.CORSAllowedOrigins, ",")
	mw := middleware.New(jwtService, authzService, origins)

	// Handlers
	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userService)
	policyHandler := handler.NewPolicyHandler(authzService)

	// Build router
	return router.SetupRouter(authHandler, userHandler, policyHandler, mw)
}
