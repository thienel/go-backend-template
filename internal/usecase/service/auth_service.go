package service

import (
	"context"

	"github.com/thienel/go-backend-template/internal/interface/api/dto"
)

// AuthService defines authentication service interface
type AuthService interface {
	Login(ctx context.Context, username, password string) (*dto.LoginResponse, error)
	Logout(ctx context.Context) error
}
