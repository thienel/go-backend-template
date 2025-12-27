package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thienel/tlog"
	"go.uber.org/zap"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/pkg/config"
	"github.com/thienel/go-backend-template/pkg/cookie"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/jwt"
	"github.com/thienel/go-backend-template/pkg/response"
)

// Context keys
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRole     = "role"
)

// RequireAuth middleware checks for valid JWT token
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()
		if cfg == nil {
			tlog.Error("RequireAuth failed: config is nil")
			response.WriteErrorResponse(c, apperror.ErrInternalServerError)
			c.Abort()
			return
		}

		var tokenString string
		var claims *jwt.Claims

		// Try Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Fall back to cookie
		if tokenString == "" {
			if token, err := cookie.GetAuthCookie(c, cfg.Cookie.Name); err == nil && token != "" {
				tokenString = token
			}
		}

		// No token found - try refresh
		if tokenString == "" {
			tlog.Debug("No access token, trying refresh", zap.String("path", c.Request.URL.Path))
			refreshedClaims := tryRefreshTokens(c, cfg)
			if refreshedClaims == nil {
				clearAllAuthCookies(c, cfg)
				response.WriteErrorResponse(c, apperror.ErrUnauthorized)
				c.Abort()
				return
			}
			claims = refreshedClaims
		} else {
			var err error
			claims, err = jwt.ValidateToken(tokenString)
			if err != nil {
				tlog.Debug("Token validation failed, trying refresh", zap.String("path", c.Request.URL.Path), zap.Error(err))
				refreshedClaims := tryRefreshTokens(c, cfg)
				if refreshedClaims != nil {
					claims = refreshedClaims
				} else {
					clearAllAuthCookies(c, cfg)
					response.WriteErrorResponse(c, apperror.ErrUnauthorized)
					c.Abort()
					return
				}
			}
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)

		c.Next()
	}
}

func tryRefreshTokens(c *gin.Context, cfg *config.Config) *jwt.Claims {
	refreshToken, err := cookie.GetRefreshCookie(c, cfg.Cookie.RefreshName)
	if err != nil || refreshToken == "" {
		return nil
	}

	claims, err := jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil
	}

	newAccessToken, err := jwt.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return nil
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return nil
	}

	cookie.SetAuthCookie(c, newAccessToken, jwt.GetAccessExpiryTime(), &cfg.Cookie)
	cookie.SetRefreshCookie(c, newRefreshToken, jwt.GetRefreshExpiryTime(), &cfg.Cookie)

	return claims
}

func clearAllAuthCookies(c *gin.Context, cfg *config.Config) {
	cookie.ClearAuthCookie(c, &cfg.Cookie)
	cookie.ClearRefreshCookie(c, &cfg.Cookie)
}

// RequireAdmin checks if user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.WriteErrorResponse(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || (roleStr != entity.UserRoleAdmin && roleStr != entity.UserRoleSystemAdmin) {
			response.WriteErrorResponse(c, apperror.ErrForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID returns user ID from context
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// GetUsername returns username from context
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(ContextKeyUsername); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetRole returns role from context
func GetRole(c *gin.Context) string {
	if role, exists := c.Get(ContextKeyRole); exists {
		if r, ok := role.(string); ok {
			return r
		}
	}
	return ""
}
