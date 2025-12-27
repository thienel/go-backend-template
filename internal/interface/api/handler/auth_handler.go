package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/interface/api/dto"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/internal/usecase/service"
	"github.com/thienel/go-backend-template/pkg/config"
	"github.com/thienel/go-backend-template/pkg/cookie"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/jwt"
	"github.com/thienel/go-backend-template/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	user, accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	cfg := config.GetConfig()
	cookie.SetAuthCookie(c, accessToken, jwt.GetAccessExpiryTime(), &cfg.Cookie)
	cookie.SetRefreshCookie(c, refreshToken, jwt.GetRefreshExpiryTime(), &cfg.Cookie)

	response.OK(c, dto.LoginResponse{
		User:         toAuthUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "Đăng nhập thành công")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	cfg := config.GetConfig()
	cookie.ClearAuthCookie(c, &cfg.Cookie)
	cookie.ClearRefreshCookie(c, &cfg.Cookie)
	response.OK[any](c, nil, "Đăng xuất thành công")
}

// GetMe returns current user info
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}
	response.OK(c, toAuthUserResponse(user), "")
}

func toAuthUserResponse(user *entity.User) dto.UserResponse {
	resp := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.DeletedAt.Valid {
		resp.DeletedAt = &user.DeletedAt.Time
	}
	return resp
}
