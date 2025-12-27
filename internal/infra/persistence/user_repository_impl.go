package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/thienel/go-backend-template/internal/domain/entity"
	"github.com/thienel/go-backend-template/internal/domain/repository"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/query"
)

// userAllowedFields defines filterable/sortable fields for security
var userAllowedFields = map[string]bool{
	"id":         true,
	"username":   true,
	"email":      true,
	"role":       true,
	"status":     true,
	"created_at": true,
	"updated_at": true,
}

type userRepositoryImpl struct {
	*BaseRepositoryImpl[entity.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	base := NewBaseRepository[entity.User](db, userAllowedFields, "người dùng")
	return &userRepositoryImpl{BaseRepositoryImpl: base}
}

func (r *userRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return &user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return &user, nil
}

func (r *userRepositoryImpl) FindByUsernameIncludingDeleted(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Unscoped().Where("username = ?", username).First(&user).Error; err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return &user, nil
}

func (r *userRepositoryImpl) FindByEmailIncludingDeleted(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.DB.WithContext(ctx).Unscoped().Where("email = ?", email).First(&user).Error; err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return &user, nil
}

func (r *userRepositoryImpl) Restore(ctx context.Context, id uint) error {
	if err := r.DB.WithContext(ctx).Unscoped().Model(&entity.User{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		return apperror.ErrInternalServerError.WithMessage("Không thể khôi phục người dùng").WithError(err)
	}
	return nil
}

func (r *userRepositoryImpl) ListWithQuery(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	q := r.DB.WithContext(ctx).Model(&entity.User{})

	// Handle special 'search' filter
	if searchFilter, ok := opts.Filters["search"]; ok {
		searchValue := searchFilter.Value.(string)
		q = q.Where("username ILIKE ? OR email ILIKE ?", "%"+searchValue+"%", "%"+searchValue+"%")
		delete(opts.Filters, "search")
	}

	// Apply filters
	q = q.Scopes(query.ApplyFilters(opts, r.AllowedFields))

	// Count total
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, wrapListError(err, "người dùng")
	}

	// Apply sort and pagination
	if err := q.Scopes(
		query.ApplySort(opts, r.AllowedFields),
		query.ApplyDefaultSort(opts, "created_at", true),
	).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, wrapListError(err, "người dùng")
	}

	return users, total, nil
}
