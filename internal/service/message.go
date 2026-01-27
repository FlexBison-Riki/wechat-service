package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wechat-service/internal/model"
	"wechat-service/pkg/cache"
	"wechat-service/pkg/logger"
)

// MessageService handles message business logic
type MessageService struct {
	cache  cache.Cache
	logger *logger.Logger
	prefix string
}

// NewMessageService creates a new MessageService
func NewMessageService(cache cache.Cache) *MessageService {
	return &MessageService{
		cache:  cache,
		logger: logger.New(),
		prefix: "msg:",
	}
}

// Save saves a message to cache/database
func (s *MessageService) Save(msg *model.Message) error {
	key := s.prefix + fmt.Sprintf("msg:%d", msg.MsgID)
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Store for 30 days
	if err := s.cache.Set(key, data, 30*24*time.Hour); err != nil {
		return fmt.Errorf("failed to cache message: %w", err)
	}

	// Add to user's message list
	listKey := s.prefix + fmt.Sprintf("user:%s:msgs", msg.FromUser)
	if err := s.addToList(listKey, msg.MsgID, 30*24*time.Hour); err != nil {
		s.logger.Warnf("Failed to add to message list: %v", err)
	}

	s.logger.Infof("Message saved: id=%d, type=%s", msg.MsgID, msg.MsgType)
	return nil
}

// GetByID retrieves a message by ID
func (s *MessageService) GetByID(id int64) (*model.Message, error) {
	key := s.prefix + fmt.Sprintf("msg:%d", id)
	data, err := s.cache.Get(key)
	if err != nil {
		return nil, fmt.Errorf("message not found: %d", id)
	}

	var msg model.Message
	if err := json.Unmarshal(data.([]byte), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// List retrieves messages for a user with pagination
func (s *MessageService) List(openid string, limit, offset int) ([]*model.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	listKey := s.prefix + fmt.Sprintf("user:%s:msgs", openid)
	ids, err := s.getListRange(listKey, offset, offset+limit-1)
	if err != nil {
		// Return empty list if no messages
		return []*model.Message{}, nil
	}

	messages := make([]*model.Message, 0, len(ids))
	for _, id := range ids {
		msg, err := s.GetByID(id)
		if err != nil {
			s.logger.Warnf("Failed to get message %d: %v", id, err)
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// Count returns the total number of messages for a user
func (s *MessageService) Count(openid string) (int, error) {
	listKey := s.prefix + fmt.Sprintf("user:%s:msgs", openid)
	count, err := s.getListLength(listKey)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Delete removes a message
func (s *MessageService) Delete(id int64) error {
	key := s.prefix + fmt.Sprintf("msg:%d", id)
	return s.cache.Set(key, nil, 0)
}

// addToList adds an ID to a sorted list
func (s *MessageService) addToList(key string, id int64, ttl time.Duration) error {
	// Use Redis sorted set with timestamp as score
	score := float64(time.Now().UnixNano())
	// For simplicity, using JSON string - in production, use Redis sorted set directly
	return nil // Placeholder
}

// getListRange retrieves IDs from a list range
func (s *MessageService) getListRange(key string, start, stop int) ([]int64, error) {
	// Placeholder - in production, use Redis sorted set
	return []int64{}, nil
}

// getListLength returns the length of a list
func (s *MessageService) getListLength(key string) (int, error) {
	// Placeholder - in production, use Redis
	return 0, nil
}

// SaveBatch saves multiple messages
func (s *MessageService) SaveBatch(ctx context.Context, messages []*model.Message) error {
	for _, msg := range messages {
		if err := s.Save(msg); err != nil {
			return err
		}
	}
	return nil
}
