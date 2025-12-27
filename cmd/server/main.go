package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/thienel/tlog"
	"go.uber.org/zap"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/infra/database"
	"github.com/thienel/go-backend-template/internal/infra/persistence"
	"github.com/thienel/go-backend-template/internal/interface/api/handler"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/internal/interface/api/router"
	"github.com/thienel/go-backend-template/internal/usecase/service/serviceimpl"
	"github.com/thienel/go-backend-template/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize tlog
	err = tlog.Init(tlog.Config{
		Environment:   cfg.Server.Env,
		Level:         cfg.Log.Level,
		AppName:       cfg.Server.ServiceName,
		EnableConsole: cfg.Log.EnableConsole,
		EnableFile:    true,
		FilePath:      cfg.Log.FilePath,
		MaxSizeMB:     cfg.Log.MaxSizeMB,
		MaxBackups:    cfg.Log.MaxBackups,
		MaxAgeDays:    cfg.Log.MaxAgeDays,
		Compress:      cfg.Log.Compress,
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer tlog.Sync()

	tlog.Info("Starting server",
		zap.String("service", cfg.Server.ServiceName),
		zap.String("version", cfg.Server.Version),
		zap.String("env", cfg.Server.Env),
	)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		tlog.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// Auto migrate
	if err := database.AutoMigrate(&entity.User{}); err != nil {
		tlog.Fatal("Failed to run auto migration", zap.Error(err))
	}
	tlog.Info("Database migration completed")

	// Initialize repositories
	db := database.GetDB()
	userRepo := persistence.NewUserRepository(db)

	// Initialize services
	jwtService := serviceimpl.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiryMinutes,
		cfg.JWT.RefreshExpiryHours,
	)
	authService := serviceimpl.NewAuthService(userRepo, jwtService)
	userService := serviceimpl.NewUserService(userRepo)

	// Initialize middleware
	origins := strings.Join(cfg.CORSAllowedOrigins, ",")
	mw := middleware.New(jwtService, origins)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, userService)
	userHandler := handler.NewUserHandler(userService)

	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup router
	engine := router.SetupRouter(authHandler, userHandler, mw)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		tlog.Info("Server starting",
			zap.String("port", cfg.Server.Port),
			zap.String("env", cfg.Server.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			tlog.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	tlog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		tlog.Fatal("Server forced to shutdown", zap.Error(err))
	}

	tlog.Info("Server exited gracefully")
}
