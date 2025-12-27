package repository

import (
	"context"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/pkg/query"
)

// UserFilter defines filter options for user queries
type UserFilter struct {
	Role   string
	Status string
	Search string
}

// UserRepository extends BaseRepository for User entity
type UserRepository interface {
	BaseRepository[entity.User]

	// Additional user-specific methods
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByUsernameIncludingDeleted(ctx context.Context, username string) (*entity.User, error)
	FindByEmailIncludingDeleted(ctx context.Context, email string) (*entity.User, error)
	Restore(ctx context.Context, id uint) error
	ListWithFilter(ctx context.Context, offset, limit int, filter UserFilter) ([]entity.User, int64, error)
	ListWithQuery(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]entity.User, int64, error)
}
