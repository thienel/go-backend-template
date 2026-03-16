package service

import (
	"context"

	"github.com/thienel/go-backend-template/internal/ent"
	"github.com/thienel/go-backend-template/pkg/query"
)

// CreateUserCommand represents the command to create a user
type CreateUserCommand struct {
	Username string
	Email    string
	Password string
	Role     string
}

// UpdateUserCommand represents the command to update a user
type UpdateUserCommand struct {
	ID       uint
	Username string
	Email    string
	Role     string
	Status   string
}

// UserService defines the user service interface
type UserService interface {
	// CRUD
	Create(ctx context.Context, cmd CreateUserCommand) (*ent.User, error)
	GetByID(ctx context.Context, id uint) (*ent.User, error)
	Update(ctx context.Context, cmd UpdateUserCommand) (*ent.User, error)
	Delete(ctx context.Context, id uint) error

	// Query
	List(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]*ent.User, int64, error)
}
