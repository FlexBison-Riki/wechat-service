package service

import (
	"fmt"
	"time"

	"wechat-service/internal/repository"
	"wechat-service/pkg/logger"

	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// EventService handles event business logic
type EventService struct {
	userRepo *repository.UserRepository
	log      *logger.Logger
}

// NewEventService creates a new event service
func NewEventService(userRepo *repository.UserRepository, log *logger.Logger) *EventService {
	return &EventService{
		userRepo: userRepo,
		log:      log,
	}
}

// OnEvent handles all events
func (s *EventService) OnEvent(msg *message.MixMessage) *message.Reply {
	switch msg.Event {
	case message.EventSubscribe:
		return s.OnSubscribe(msg)
	case message.EventUnsubscribe:
		return s.OnUnsubscribe(msg)
	case message.EventScan:
		return s.OnScan(msg)
	case message.EventClick:
		return s.OnClick(msg)
	case message.EventView:
		return s.OnView(msg)
	case message.EventLocation:
		return s.OnLocation(msg)
	case message.EventMasssendJobFinish:
		return s.OnMassSendJobFinish(msg)
	default:
		s.log.Warn("Unknown event type", "event", msg.Event)
		return nil
	}
}

// OnSubscribe handles subscribe events
func (s *EventService) OnSubscribe(msg *message.MixMessage) *message.Reply {
	s.log.Info("User subscribed", "openid", msg.FromUserName)

	// Save user if repository available
	if s.userRepo != nil {
		user := &repository.User{
			OpenID:        msg.FromUserName,
			Subscribe:     1,
			SubscribeTime: time.Now(),
		}
		s.userRepo.Save(user)
	}

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: "欢迎关注我们的服务号！\n发送任意内容开始体验。",
		},
	}
}

// OnUnsubscribe handles unsubscribe events
func (s *EventService) OnUnsubscribe(msg *message.MixMessage) *message.Reply {
	s.log.Info("User unsubscribed", "openid", msg.FromUserName)

	// Update user subscription status
	if s.userRepo != nil {
		user, _ := s.userRepo.GetByOpenID(msg.FromUserName)
		if user != nil {
			user.Subscribe = 0
			s.userRepo.Save(user)
		}
	}

	return nil
}

// OnScan handles QR code scan events
func (s *EventService) OnScan(msg *message.MixMessage) *message.Reply {
	s.log.Info("QR code scanned",
		"openid", msg.FromUserName,
		"event_key", msg.EventKey,
		"ticket", msg.Ticket,
	)

	// Handle different scan scenarios
	return nil
}

// OnClick handles menu click events
func (s *EventService) OnClick(msg *message.MixMessage) *message.Reply {
	s.log.Info("Menu clicked",
		"openid", msg.FromUserName,
		"key", msg.EventKey,
	)

	// Route to specific handlers based on key
	switch msg.EventKey {
	case "V1001_TODAY_MUSIC":
		return s.handleTodayMusic(msg)
	case "V1001_HELP":
		return s.handleHelp(msg)
	default:
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: &message.Text{
				Content: fmt.Sprintf("您点击了: %s", msg.EventKey),
			},
		}
	}
}

// OnView handles menu view events
func (s *EventService) OnView(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Menu viewed", "url", msg.EventKey)
	return nil
}

// OnLocation handles location events
func (s *EventService) OnLocation(msg *message.MixMessage) *message.Reply {
	s.log.Debug("Location updated",
		"openid", msg.FromUserName,
		"lat", msg.Latitude,
		"lng", msg.Longitude,
	)
	return nil
}

// OnMassSendJobFinish handles mass send completion events
func (s *EventService) OnMassSendJobFinish(msg *message.MixMessage) *message.Reply {
	s.log.Info("Mass send job finished",
		"status", msg.Status,
		"total_count", msg.TotalCount,
		"filter_count", msg.FilterCount,
		"sent_count", msg.SentCount,
		"error_count", msg.ErrorCount,
	)
	return nil
}

// handleTodayMusic handles "Today's Music" menu click
func (s *EventService) handleTodayMusic(msg *message.MixMessage) *message.Reply {
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: "🎵 今日推荐音乐功能开发中...\n敬请期待！",
		},
	}
}

// handleHelp handles "Help" menu click
func (s *EventService) handleHelp(msg *message.MixMessage) *message.Reply {
	helpText := `🤖 服务号使用指南

📋 菜单说明：
- 点击菜单按钮可触发相应功能
- 部分功能需要授权后使用

💬 消息类型支持：
- 文本消息
- 图片消息
- 语音消息
- 地理位置

❓ 问题反馈：
请发送消息给管理员`

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: helpText,
		},
	}
}
