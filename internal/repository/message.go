package repository

import (
	"context"
	"sync"

	"wechat-service/internal/model"
)

// MessageRepository simple in-memory message repository
type MessageRepository struct {
	mu       sync.RWMutex
	messages map[int64]*model.Message
}

// NewMessageRepository creates a new in-memory message repository
func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		messages: make(map[int64]*model.Message),
	}
}

// Save saves a message
func (r *MessageRepository) Save(ctx context.Context, msg *model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	msg.ID = int64(len(r.messages) + 1)
	r.messages[msg.MsgID] = msg
	return nil
}

// GetByMsgID retrieves a message by WeChat msg_id
func (r *MessageRepository) GetByMsgID(ctx context.Context, msgID int64) (*model.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msg, ok := r.messages[msgID]
	if !ok {
		return nil, ErrMessageNotFound
	}
	return msg, nil
}

// Exists checks if a message exists
func (r *MessageRepository) Exists(ctx context.Context, msgID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.messages[msgID]
	return ok
}

// ListByUser retrieves messages for a user
func (r *MessageRepository) ListByUser(ctx context.Context, openid string, limit, offset int) ([]*model.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var msgs []*model.Message
	count := 0
	for _, msg := range r.messages {
		if msg.FromUser == openid || msg.ToUser == openid {
			if count >= offset && len(msgs) < limit {
				msgs = append(msgs, msg)
			}
			count++
		}
	}
	return msgs, nil
}

// Count returns total message count
func (r *MessageRepository) Count(ctx context.Context) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.messages)
}

// ErrMessageNotFound error for message not found
var ErrMessageNotFound = &RepositoryError{Message: "message not found"}
