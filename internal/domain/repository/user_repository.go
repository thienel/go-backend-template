package repository

import (
	"context"

	"github.com/thienel/go-backend-template/internal/ent"
	"github.com/thienel/go-backend-template/pkg/query"
)

// UserRepository extends BaseRepository for User entity
type UserRepository interface {
	BaseRepository[ent.User]

	// Additional user-specific methods
	FindByUsername(ctx context.Context, username string) (*ent.User, error)
	FindByEmail(ctx context.Context, email string) (*ent.User, error)
	FindByUsernameIncludingDeleted(ctx context.Context, username string) (*ent.User, error)
	FindByEmailIncludingDeleted(ctx context.Context, email string) (*ent.User, error)
	Restore(ctx context.Context, id uint) error

	// ListWithQuery supports search filter across multiple fields
	ListWithQuery(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]*ent.User, int64, error)
}
