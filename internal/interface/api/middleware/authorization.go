package middleware

import (
	"github.com/gin-gonic/gin"

	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/response"
)

// Authorize returns Casbin authorization middleware
// Must be used AFTER Auth() middleware (requires JWT claims in context)
func (m *Middleware) Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetUserClaims(c)
		if claims == nil {
			response.WriteErrorResponse(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		// Get request info
		role := claims.Role
		path := c.FullPath() // e.g. "/api/users/:id"
		method := c.Request.Method

		// If FullPath is empty (no matching route), use the raw URL path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Enforce policy via Casbin
		allowed, err := m.authzService.Enforce(role, path, method)
		if err != nil {
			response.WriteErrorResponse(c, apperror.ErrInternalServerError.WithMessage("Lỗi kiểm tra quyền truy cập").WithError(err))
			c.Abort()
			return
		}

		if !allowed {
			response.WriteErrorResponse(c, apperror.ErrForbidden.WithMessage("Bạn không có quyền thực hiện hành động này"))
			c.Abort()
			return
		}

		c.Next()
	}
}
