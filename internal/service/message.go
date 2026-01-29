package service

import (
	"time"

	"wechat-service/internal/repository"
	"wechat-service/pkg/logger"

	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// MessageService handles message business logic
type MessageService struct {
	repo *repository.MessageRepository
	log  *logger.Logger
}

// NewMessageService creates a new message service
func NewMessageService(repo *repository.MessageRepository, log *logger.Logger) *MessageService {
	return &MessageService{
		repo: repo,
		log:  log,
	}
}

// OnTextMessage handles text messages
func (s *MessageService) OnTextMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing text message", "content", msg.Content)

	// Simple echo response
	// In production, implement business logic here
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: msg.Content,
		},
	}
}

// OnImageMessage handles image messages
func (s *MessageService) OnImageMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing image message", "media_id", msg.MediaID)

	// Echo image
	return &message.Reply{
		MsgType: message.MsgTypeImage,
		MsgData: &message.Image{
			MediaID: msg.MediaID,
		},
	}
}

// OnVoiceMessage handles voice messages
func (s *MessageService) OnVoiceMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing voice message", "media_id", msg.MediaID)
	return nil
}

// OnVideoMessage handles video messages
func (s *MessageService) OnVideoMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing video message", "media_id", msg.MediaID)
	return nil
}

// OnLocationMessage handles location messages
func (s *MessageService) OnLocationMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing location message",
		"x", msg.LocationX,
		"y", msg.LocationY,
		"label", msg.Label,
	)
	return nil
}

// OnLinkMessage handles link messages
func (s *MessageService) OnLinkMessage(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Processing link message", "title", msg.Title)
	return nil
}

// SaveMessage saves a message to repository
func (s *MessageService) SaveMessage(msg *message.MixMessage) error {
	if s.repo == nil {
		return nil
	}

	repoMsg := &repository.Message{
		MsgID:        msg.MsgID,
		FromUser:     msg.FromUserName,
		ToUser:       msg.ToUserName,
		MsgType:      msg.MsgType,
		Content:      msg.Content,
		MediaID:      msg.MediaID,
		PicURL:       msg.PicURL,
		Format:       msg.Format,
		ThumbMediaID: msg.ThumbMediaID,
		LocationX:    msg.LocationX,
		LocationY:    msg.LocationY,
		Scale:        msg.Scale,
		Label:        msg.Label,
		Title:        msg.Title,
		Description:  msg.Description,
		URL:          msg.URL,
		Event:        msg.Event,
		EventKey:     msg.EventKey,
		Ticket:       msg.Ticket,
		Latitude:     msg.Latitude,
		Longitude:    msg.Longitude,
		Precision:    msg.Precision,
		MsgDataID:    msg.MsgDataID,
		Idx:          msg.Idx,
		CreatedAt:    time.Now(),
	}

	return s.repo.Save(repoMsg)
}

// GetRecentMessages retrieves recent messages
func (s *MessageService) GetRecentMessages(limit int) ([]*repository.Message, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.GetRecent(limit)
}

// GetMessageStats returns message statistics
func (s *MessageService) GetMessageStats() map[string]int {
	if s.repo == nil {
		return nil
	}

	stats := make(map[string]int)
	types := []string{
		message.MsgTypeText,
		message.MsgTypeImage,
		message.MsgTypeVoice,
		message.MsgTypeVideo,
		message.MsgTypeLocation,
		message.MsgTypeLink,
	}

	for _, msgType := range types {
		stats[msgType] = s.repo.CountByType(msgType)
	}

	stats["total"] = s.repo.Count()
	return stats
}
