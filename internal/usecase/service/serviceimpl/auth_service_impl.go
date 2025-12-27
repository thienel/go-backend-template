package serviceimpl

import (
	"context"

	"github.com/thienel/tlog"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/domain/repository"
	"github.com/thienel/go-backend-template/internal/interface/api/dto"
	"github.com/thienel/go-backend-template/internal/usecase/service"
	apperror "github.com/thienel/go-backend-template/pkg/error"
)

type authServiceImpl struct {
	userRepo   repository.UserRepository
	jwtService service.JWTService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, jwtService service.JWTService) service.AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, username, password string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		tlog.Debug("Login failed: user not found", zap.String("username", username))
		return nil, apperror.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		tlog.Debug("Login failed: invalid password", zap.String("username", username))
		return nil, apperror.ErrInvalidCredentials
	}

	if user.Status != entity.UserStatusActive {
		tlog.Debug("Login failed: user inactive", zap.String("username", username))
		return nil, apperror.ErrForbidden.WithMessage("Tài khoản đã bị vô hiệu hóa")
	}

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, apperror.ErrInternalServerError.WithMessage("Không thể tạo access token").WithError(err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, apperror.ErrInternalServerError.WithMessage("Không thể tạo refresh token").WithError(err)
	}

	tlog.Info("User logged in", zap.Uint("user_id", user.ID), zap.String("username", user.Username))

	return &dto.LoginResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authServiceImpl) Logout(ctx context.Context) error {
	// For stateless JWT, logout is handled at the handler level by clearing cookies
	// If you need blacklist/revocation, implement it here with Redis
	return nil
}
