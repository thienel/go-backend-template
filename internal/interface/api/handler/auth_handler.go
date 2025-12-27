package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/interface/api/dto"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/internal/usecase/service"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/response"
)

// AuthHandler interface
type AuthHandler interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
	GetMe(c *gin.Context)
}

type authHandlerImpl struct {
	authService service.AuthService
	userService service.UserService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService, userService service.UserService) AuthHandler {
	return &authHandlerImpl{
		authService: authService,
		userService: userService,
	}
}

func (h *authHandlerImpl) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	loginResp, err := h.authService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	response.OK(c, loginResp, "Đăng nhập thành công")
}

func (h *authHandlerImpl) Logout(c *gin.Context) {
	if err := h.authService.Logout(c.Request.Context()); err != nil {
		response.WriteErrorResponse(c, err)
		return
	}
	response.OK[any](c, nil, "Đăng xuất thành công")
}

func (h *authHandlerImpl) GetMe(c *gin.Context) {
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
