package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/domain/repository"
	"github.com/thienel/go-backend-template/internal/interface/api/dto"
	"github.com/thienel/go-backend-template/internal/interface/api/middleware"
	"github.com/thienel/go-backend-template/pkg/config"
	"github.com/thienel/go-backend-template/pkg/cookie"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/jwt"
	"github.com/thienel/go-backend-template/pkg/query"
	"github.com/thienel/go-backend-template/pkg/response"
)

var userAllowedFields = map[string]bool{
	"id":         true,
	"username":   true,
	"email":      true,
	"role":       true,
	"status":     true,
	"created_at": true,
	"search":     true,
}

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	user, err := h.userRepo.FindByUsername(c.Request.Context(), req.Username)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInvalidCredentials)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.WriteErrorResponse(c, apperror.ErrInvalidCredentials)
		return
	}

	if user.Status != entity.UserStatusActive {
		response.WriteErrorResponse(c, apperror.ErrForbidden.WithMessage("Tài khoản đã bị vô hiệu hóa"))
		return
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError)
		return
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError)
		return
	}

	cfg := config.GetConfig()
	cookie.SetAuthCookie(c, accessToken, jwt.GetAccessExpiryTime(), &cfg.Cookie)
	cookie.SetRefreshCookie(c, refreshToken, jwt.GetRefreshExpiryTime(), &cfg.Cookie)

	response.OK(c, dto.LoginResponse{
		User:         toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, "Đăng nhập thành công")
}

// Logout handles user logout
func (h *UserHandler) Logout(c *gin.Context) {
	cfg := config.GetConfig()
	cookie.ClearAuthCookie(c, &cfg.Cookie)
	cookie.ClearRefreshCookie(c, &cfg.Cookie)
	response.OK[any](c, nil, "Đăng xuất thành công")
}

// GetMe returns current user info
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userRepo.FindByID(c.Request.Context(), userID)
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}
	response.OK(c, toUserResponse(user), "")
}

// List lists all users
func (h *UserHandler) List(c *gin.Context) {
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	offset, limit := query.GetPagination(params, 20)
	opts := query.ParseQueryParams(params, userAllowedFields)

	users, total, err := h.userRepo.ListWithQuery(c.Request.Context(), offset, limit, opts)
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	items := make([]dto.UserResponse, len(users))
	for i, u := range users {
		items[i] = toUserResponse(&u)
	}

	page := (offset / limit) + 1
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response.OK(c, dto.ListResponse[dto.UserResponse]{
		Items:      items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, "")
}

// GetByID returns user by ID
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("ID không hợp lệ"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	response.OK(c, toUserResponse(user), "")
}

// Create creates a new user
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	// Check username exists
	if _, err := h.userRepo.FindByUsernameIncludingDeleted(c.Request.Context(), req.Username); err == nil {
		response.WriteErrorResponse(c, apperror.ErrUsernameExists)
		return
	}

	// Check email exists
	if _, err := h.userRepo.FindByEmailIncludingDeleted(c.Request.Context(), req.Email); err == nil {
		response.WriteErrorResponse(c, apperror.ErrEmailExists)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrInternalServerError)
		return
	}

	role := entity.UserRoleUser
	if req.Role != "" {
		if !entity.IsValidUserRole(req.Role) {
			response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Role không hợp lệ"))
			return
		}
		role = req.Role
	}

	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
		Status:   entity.UserStatusActive,
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	response.Created(c, toUserResponse(user), "Tạo người dùng thành công")
}

// Update updates a user
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("ID không hợp lệ"))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Dữ liệu không hợp lệ"))
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	if req.Username != "" && req.Username != user.Username {
		if _, err := h.userRepo.FindByUsernameIncludingDeleted(c.Request.Context(), req.Username); err == nil {
			response.WriteErrorResponse(c, apperror.ErrUsernameExists)
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		if _, err := h.userRepo.FindByEmailIncludingDeleted(c.Request.Context(), req.Email); err == nil {
			response.WriteErrorResponse(c, apperror.ErrEmailExists)
			return
		}
		user.Email = req.Email
	}

	if req.Role != "" {
		if !entity.IsValidUserRole(req.Role) {
			response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Role không hợp lệ"))
			return
		}
		user.Role = req.Role
	}
	if req.Status != "" {
		if !entity.IsValidUserStatus(req.Status) {
			response.WriteErrorResponse(c, apperror.ErrValidation.WithMessage("Status không hợp lệ"))
			return
		}
		user.Status = req.Status
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	response.OK(c, toUserResponse(user), "Cập nhật thành công")
}

// Delete soft-deletes a user
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.WriteErrorResponse(c, apperror.ErrBadRequest.WithMessage("ID không hợp lệ"))
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		response.WriteErrorResponse(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func toUserResponse(user *entity.User) dto.UserResponse {
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
