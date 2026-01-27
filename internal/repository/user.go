package repository

import (
	"context"
	"sync"
	"time"

	"wechat-service/internal/model"
)

// UserRepository simple in-memory user repository
type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*model.User
}

// NewUserRepository creates a new in-memory user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*model.User),
	}
}

// Save saves or updates a user
func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user.UpdatedAt = time.Now()
	r.users[user.OpenID] = user
	return nil
}

// GetByOpenID retrieves a user by OpenID
func (r *UserRepository) GetByOpenID(ctx context.Context, openid string) (*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[openid]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Exists checks if a user exists
func (r *UserRepository) Exists(ctx context.Context, openid string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.users[openid]
	return ok
}

// UpdateStatus updates user's subscription status
func (r *UserRepository) UpdateStatus(ctx context.Context, openid string, subscribed bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[openid]
	if !ok {
		return ErrUserNotFound
	}
	user.SubscribeStatus = 1
	if !subscribed {
		user.SubscribeStatus = 0
		now := time.Now()
		user.UnsubscribeTime = &now
	}
	user.UpdatedAt = time.Now()
	return nil
}

// Subscribe marks a user as subscribed
func (r *UserRepository) Subscribe(ctx context.Context, user *model.User) error {
	user.SubscribeStatus = 1
	user.SubscribeTime = time.Now()
	return r.Save(ctx, user)
}

// Unsubscribe marks a user as unsubscribed
func (r *UserRepository) Unsubscribe(ctx context.Context, openid string) error {
	return r.UpdateStatus(ctx, openid, false)
}

// Delete removes a user (GDPR compliance)
func (r *UserRepository) Delete(ctx context.Context, openid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[openid]; !ok {
		return ErrUserNotFound
	}
	delete(r.users, openid)
	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*model.User, 0, limit)
	count := 0
	for _, user := range r.users {
		if count >= offset && len(users) < limit {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}

// Count returns total user count
func (r *UserRepository) Count(ctx context.Context) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}

// ErrUserNotFound error for user not found
var ErrUserNotFound = &RepositoryError{Message: "user not found"}

type RepositoryError struct {
	Message string
}

func (e *RepositoryError) Error() string {
	return e.Message
}
