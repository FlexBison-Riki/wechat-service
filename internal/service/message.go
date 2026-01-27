package service

import (
	"context"

	"wechat-service/internal/model"
	"wechat-service/internal/repository"
)

// MessageService handles message business logic
type MessageService struct {
	repo *repository.MessageRepository
}

// NewMessageService creates a new MessageService
func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

// Save saves a message
func (s *MessageService) Save(msg *model.Message) error {
	return s.repo.Save(context.Background(), msg)
}

// GetByMsgID retrieves a message by msg_id
func (s *MessageService) GetByMsgID(msgID int64) (*model.Message, error) {
	return s.repo.GetByMsgID(context.Background(), msgID)
}

// Exists checks if a message exists
func (s *MessageService) Exists(msgID int64) bool {
	return s.repo.Exists(context.Background(), msgID)
}

// ListByUser retrieves messages for a user
func (s *MessageService) ListByUser(openid string, limit, offset int) ([]*model.Message, error) {
	return s.repo.ListByUser(context.Background(), openid, limit, offset)
}

// Count returns total message count
func (s *MessageService) Count() int {
	return s.repo.Count(context.Background())
}
