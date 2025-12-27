package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/thienel/go-backend-template/pkg/config"
)

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
	cfg := config.GetConfig()
	origins := []string{"*"}
	if cfg != nil && len(cfg.CORSAllowedOrigins) > 0 {
		origins = cfg.CORSAllowedOrigins
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// Recovery returns a custom recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"is_success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Đã xảy ra lỗi máy chủ nội bộ",
			},
		})
	})
}
