package service

import (
	"wechat-service/internal/repository"

	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// MessageService handles message business logic
type MessageService struct {
	repo *repository.MessageRepository
}

// NewMessageService creates a new message service
func NewMessageService(repo *repository.MessageRepository) *MessageService {
	return &MessageService{
		repo: repo,
	}
}

// OnTextMessage handles text messages - SDK calls this
func (s *MessageService) OnTextMessage(msg *message.MixMessage) *message.Reply {
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: message.NewText(string(msg.Content)),
	}
}

// OnImageMessage handles image messages
func (s *MessageService) OnImageMessage(msg *message.MixMessage) *message.Reply {
	return &message.Reply{
		MsgType: message.MsgTypeImage,
		MsgData: message.NewImage(string(msg.MediaID)),
	}
}

// OnVoiceMessage handles voice messages
func (s *MessageService) OnVoiceMessage(msg *message.MixMessage) *message.Reply {
	return nil
}

// OnVideoMessage handles video messages
func (s *MessageService) OnVideoMessage(msg *message.MixMessage) *message.Reply {
	return nil
}

// OnLocationMessage handles location messages
func (s *MessageService) OnLocationMessage(msg *message.MixMessage) *message.Reply {
	return nil
}

// OnLinkMessage handles link messages
func (s *MessageService) OnLinkMessage(msg *message.MixMessage) *message.Reply {
	return nil
}

// SaveMessage saves a message to repository
func (s *MessageService) SaveMessage(msg *message.MixMessage) error {
	if s.repo == nil {
		return nil
	}

	repoMsg := &repository.Message{
		MsgID: msg.MsgID,
	}

	return s.repo.Save(repoMsg)
}

// GetMessageStats returns message statistics
func (s *MessageService) GetMessageStats() map[string]int {
	if s.repo == nil {
		return nil
	}

	return map[string]int{
		"text":  s.repo.CountByType("text"),
		"image": s.repo.CountByType("image"),
		"voice": s.repo.CountByType("voice"),
		"video": s.repo.CountByType("video"),
		"total": s.repo.Count(),
	}
}
