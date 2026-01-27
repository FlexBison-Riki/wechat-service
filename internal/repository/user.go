package repository

import (
	"context"

	"wechat-service/internal/model"
)

// UserRepository interface defines user data operations
type UserRepository interface {
	Save(ctx context.Context, user *model.User) error
	GetByOpenID(ctx context.Context, openid string) (*model.User, error)
	Exists(ctx context.Context, openid string) (bool, error)
	UpdateStatus(ctx context.Context, openid string, subscribed bool) error
	UpdateLocation(ctx context.Context, openid string, lat, lng, precision float64) error
	Subscribe(ctx context.Context, user *model.User) error
	Unsubscribe(ctx context.Context, openid string) error
	Delete(ctx context.Context, openid string) error
	List(ctx context.Context, status *int, limit, offset int) ([]*model.User, error)
	Count(ctx context.Context) (int, error)
	CountSubscribed(ctx context.Context) (int, error)
}

// Compile-time check that DBUserRepository implements UserRepository
var _ UserRepository = (*DBUserRepository)(nil)
