package repository

import (
	"context"
	"time"

	"wechat-service/internal/model"
	"wechat-service/pkg/cache"
)

// UserRepository handles user data access
type UserRepository struct {
	cache cache.Cache
	ttl   time.Duration
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(cache cache.Cache) *UserRepository {
	return &UserRepository{
		cache: cache,
		ttl:   365 * 24 * time.Hour,
	}
}

// Save saves or updates a user
func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	return r.cache.Set("user:"+user.OpenID, user, r.ttl)
}

// FindByOpenID finds a user by OpenID
func (r *UserRepository) FindByOpenID(ctx context.Context, openid string) (*model.User, error) {
	data, err := r.cache.Get("user:" + openid)
	if err != nil {
		return nil, err
	}
	user, ok := data.(*model.User)
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Exists checks if a user exists
func (r *UserRepository) Exists(ctx context.Context, openid string) bool {
	_, err := r.cache.Get("user:" + openid)
	return err == nil
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, openid string) error {
	return r.cache.Set("user:"+openid, nil, 0)
}

// ListByTag retrieves users by tag (placeholder)
func (r *UserRepository) ListByTag(ctx context.Context, tagID int, limit, offset int) ([]*model.User, error) {
	return nil, nil
}

// Count returns user count (placeholder)
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

// Custom errors
var ErrUserNotFound = &RepositoryError{Message: "user not found"}

type RepositoryError struct {
	Message string
}

func (e *RepositoryError) Error() string {
	return e.Message
}
