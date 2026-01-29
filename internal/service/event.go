package service

import (
	"time"

	"wechat-service/internal/repository"

	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// EventService handles event business logic
type EventService struct {
	userRepo *repository.UserRepository
}

// NewEventService creates a new event service
func NewEventService(userRepo *repository.UserRepository) *EventService {
	return &EventService{
		userRepo: userRepo,
	}
}

// OnSubscribe handles subscribe events
func (s *EventService) OnSubscribe(msg *message.MixMessage) *message.Reply {
	if s.userRepo != nil {
		user := &repository.User{
			OpenID:        string(msg.FromUserName),
			Subscribe:     1,
			SubscribeTime: time.Now(),
		}
		s.userRepo.Save(user)
	}

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: message.CDATA("欢迎关注！发送任意内容开始体验。"),
		},
	}
}

// OnUnsubscribe handles unsubscribe events
func (s *EventService) OnUnsubscribe(msg *message.MixMessage) {
	if s.userRepo != nil {
		user, _ := s.userRepo.GetByOpenID(string(msg.FromUserName))
		if user != nil {
			user.Subscribe = 0
			s.userRepo.Save(user)
		}
	}
}

// OnScan handles QR code scan events
func (s *EventService) OnScan(msg *message.MixMessage) *message.Reply {
	return nil
}

// OnClick handles menu click events
func (s *EventService) OnClick(msg *message.MixMessage) *message.Reply {
	switch string(msg.EventKey) {
	case "V1001_HELP":
		return s.showHelp(msg)
	case "V1001_CONTACT":
		return s.showContact(msg)
	default:
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: &message.Text{
				Content: message.CDATA("您点击了: " + string(msg.EventKey)),
			},
		}
	}
}

// OnView handles menu view events
func (s *EventService) OnView(msg *message.MixMessage) {
}

// OnLocation handles location events
func (s *EventService) OnLocation(msg *message.MixMessage) {
}

// showHelp returns help information
func (s *EventService) showHelp(msg *message.MixMessage) *message.Reply {
	text := `🤖 服务号使用指南

• 发送文本消息
• 发送图片
• 点击菜单使用功能

如有疑问，请联系管理员。`

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{Content: message.CDATA(text)},
	}
}

// showContact returns contact information
func (s *EventService) showContact(msg *message.MixMessage) *message.Reply {
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: message.CDATA("📧 联系我们：请发送消息给管理员"),
		},
	}
}
