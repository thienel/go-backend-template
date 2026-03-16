package persistence

import (
	"context"

	"github.com/thienel/go-backend-template/internal/domain/repository"
	"github.com/thienel/go-backend-template/internal/ent"
	"github.com/thienel/go-backend-template/internal/ent/user"
	apperror "github.com/thienel/go-backend-template/pkg/error"
	"github.com/thienel/go-backend-template/pkg/query"
)

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
	client *ent.Client
}

// NewUserRepository creates a new user repository
func NewUserRepository(client *ent.Client) repository.UserRepository {
	return &userRepositoryImpl{client: client}
}

// --- BaseRepository Methods ---

func (r *userRepositoryImpl) Create(ctx context.Context, e *ent.User) error {
	u, err := r.client.User.Create().
		SetUsername(e.Username).
		SetEmail(e.Email).
		SetPassword(e.Password).
		SetRole(e.Role).
		SetStatus(e.Status).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return apperror.ErrConflict.WithMessage("Dữ liệu đã tồn tại (email hoặc username)").WithError(err)
		}
		return wrapCreateError(err, "người dùng")
	}
	e.ID = u.ID
	e.CreatedAt = u.CreatedAt
	e.UpdatedAt = u.UpdatedAt
	return nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id uint) (*ent.User, error) {
	u, err := r.client.User.Query().Where(user.ID(id), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.ErrNotFound.WithMessage("Không tìm thấy người dùng")
		}
		return nil, wrapFindError(err, "người dùng")
	}
	return u, nil
}

func (r *userRepositoryImpl) Update(ctx context.Context, e *ent.User) error {
	u, err := r.client.User.UpdateOneID(e.ID).
		SetUsername(e.Username).
		SetEmail(e.Email).
		SetPassword(e.Password).
		SetRole(e.Role).
		SetStatus(e.Status).
		Save(ctx)
	if err != nil {
		return wrapUpdateError(err, "người dùng")
	}
	e.UpdatedAt = u.UpdatedAt
	return nil
}

func (r *userRepositoryImpl) Delete(ctx context.Context, id uint) error {
	err := r.client.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return wrapDeleteError(err, "người dùng")
	}
	return nil
}

func (r *userRepositoryImpl) List(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]ent.User, int64, error) {
	entUsers, total, err := r.ListWithQuery(ctx, offset, limit, opts)
	if err != nil {
		return nil, 0, err
	}
	res := make([]ent.User, len(entUsers))
	for i, u := range entUsers {
		res[i] = *u
	}
	return res, total, nil
}

func (r *userRepositoryImpl) Exists(ctx context.Context, id uint) (bool, error) {
	count, err := r.client.User.Query().Where(user.ID(id), user.DeletedAtIsNil()).Count(ctx)
	if err != nil {
		return false, wrapFindError(err, "người dùng")
	}
	return count > 0, nil
}

// --- UserRepository Methods ---

func (r *userRepositoryImpl) FindByUsername(ctx context.Context, username string) (*ent.User, error) {
	u, err := r.client.User.Query().Where(user.Username(username), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return u, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*ent.User, error) {
	u, err := r.client.User.Query().Where(user.Email(email), user.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return u, nil
}

func (r *userRepositoryImpl) FindByUsernameIncludingDeleted(ctx context.Context, username string) (*ent.User, error) {
	u, err := r.client.User.Query().Where(user.Username(username)).Only(ctx)
	if err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return u, nil
}

func (r *userRepositoryImpl) FindByEmailIncludingDeleted(ctx context.Context, email string) (*ent.User, error) {
	u, err := r.client.User.Query().Where(user.Email(email)).Only(ctx)
	if err != nil {
		return nil, wrapFindError(err, "người dùng")
	}
	return u, nil
}

func (r *userRepositoryImpl) Restore(ctx context.Context, id uint) error {
	// Need to clear deleted_at.
	err := r.client.User.UpdateOneID(id).ClearDeletedAt().Exec(ctx)
	if err != nil {
		return apperror.ErrInternalServerError.WithMessage("Không thể khôi phục người dùng").WithError(err)
	}
	return nil
}

func (r *userRepositoryImpl) ListWithQuery(ctx context.Context, offset, limit int, opts query.QueryOptions) ([]*ent.User, int64, error) {
	q := r.client.User.Query().Where(user.DeletedAtIsNil())

	if searchFilter, ok := opts.Filters["search"]; ok {
		searchValue := searchFilter.Value.(string)
		q = q.Where(user.Or(
			user.UsernameContainsFold(searchValue),
			user.EmailContainsFold(searchValue),
		))
	}

	// Dynamic filters could be mapped here if needed.
	// For production we can implement a generic filter builder for Ent
	// or specific fields mapping.

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, wrapListError(err, "người dùng")
	}

	// Apply simple sort mapping
	if len(opts.Sort) > 0 {
		for _, sort := range opts.Sort {
			if !userAllowedFields[sort.Field] {
				continue
			}
			if sort.Desc {
				q = q.Order(ent.Desc(sort.Field))
			} else {
				q = q.Order(ent.Asc(sort.Field))
			}
		}
	} else {
		q = q.Order(ent.Desc(user.FieldCreatedAt))
	}

	entUsers, err := q.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, 0, wrapListError(err, "người dùng")
	}

	return entUsers, int64(total), nil
}
