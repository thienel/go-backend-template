package service

import (
	"context"

	"github.com/thienel/go-backend-template/internal/domain/entity"
)

// AuthService defines authentication service interface
type AuthService interface {
	Login(ctx context.Context, username, password string) (*entity.User, string, string, error)
	Logout(ctx context.Context) error
}
