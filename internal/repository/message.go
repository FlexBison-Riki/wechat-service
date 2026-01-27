package repository

import (
	"context"
	"time"

	"wechat-service/internal/model"
	"wechat-service/pkg/cache"
)

// MessageRepository handles message data access
type MessageRepository struct {
	cache cache.Cache
	ttl   time.Duration
}

// NewMessageRepository creates a new MessageRepository
func NewMessageRepository(cache cache.Cache) *MessageRepository {
	return &MessageRepository{
		cache: cache,
		ttl:   30 * 24 * time.Hour,
	}
}

// Save saves a message
func (r *MessageRepository) Save(ctx context.Context, msg *model.Message) error {
	key := "msg:" + fmt.Sprintf("%d", msg.MsgID)
	return r.cache.Set(key, msg, r.ttl)
}

// FindByID finds a message by ID
func (r *MessageRepository) FindByID(ctx context.Context, id int64) (*model.Message, error) {
	key := "msg:" + fmt.Sprintf("%d", id)
	data, err := r.cache.Get(key)
	if err != nil {
		return nil, ErrMessageNotFound
	}
	msg, ok := data.(*model.Message)
	if !ok {
		return nil, ErrMessageNotFound
	}
	return msg, nil
}

// ListByUser retrieves messages for a user (placeholder)
func (r *MessageRepository) ListByUser(ctx context.Context, openid string, limit, offset int) ([]*model.Message, error) {
	return nil, nil
}

// Delete removes a message
func (r *MessageRepository) Delete(ctx context.Context, id int64) error {
	key := "msg:" + fmt.Sprintf("%d", id)
	return r.cache.Set(key, nil, 0)
}

// CountByUser returns message count for a user (placeholder)
func (r *MessageRepository) CountByUser(ctx context.Context, openid string) (int, error) {
	return 0, nil
}

// Custom errors
var ErrMessageNotFound = &RepositoryError{Message: "message not found"}
