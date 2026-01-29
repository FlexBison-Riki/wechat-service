package handler

import (
	"net/http"

	"wechat-service/internal/config"
	"wechat-service/internal/service"
	"wechat-service/pkg/logger"
	"wechat-service/pkg/metrics"

	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// MessageHandler handles WeChat callbacks using the SDK
type MessageHandler struct {
	cfg      *config.Config
	oa       *officialaccount.OfficialAccount
	log      *logger.Logger
	metrics  *metrics.Metrics
	msgSvc   *service.MessageService
	eventSvc *service.EventService
}

// NewMessageHandler creates a new message handler using SDK
func NewMessageHandler(
	cfg *config.Config,
	oa *officialaccount.OfficialAccount,
	msgSvc *service.MessageService,
	eventSvc *service.EventService,
	log *logger.Logger,
	metrics *metrics.Metrics,
) *MessageHandler {
	h := &MessageHandler{
		cfg:     cfg,
		oa:      oa,
		log:     log,
		metrics: metrics,
		msgSvc:  msgSvc,
		eventSvc: eventSvc,
	}

	return h
}

// ServeHTTP handles the WeChat callback using SDK's server
func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.metrics.IncHTTPRequest("wechat", r.Method, "200")

	// Get server instance from SDK
	srv := h.oa.GetServer(r, w)

	// Set message handler
	srv.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		// Handle based on message type
		switch msg.MsgType {
		case message.MsgTypeText:
			h.metrics.IncMessageReceived("text")
			if h.msgSvc != nil {
				return h.msgSvc.OnTextMessage(msg)
			}
		case message.MsgTypeImage:
			h.metrics.IncMessageReceived("image")
			if h.msgSvc != nil {
				return h.msgSvc.OnImageMessage(msg)
			}
		case message.MsgTypeVoice:
			h.metrics.IncMessageReceived("voice")
			if h.msgSvc != nil {
				return h.msgSvc.OnVoiceMessage(msg)
			}
		case message.MsgTypeVideo:
			h.metrics.IncMessageReceived("video")
			if h.msgSvc != nil {
				return h.msgSvc.OnVideoMessage(msg)
			}
		case message.MsgTypeLocation:
			h.metrics.IncMessageReceived("location")
			if h.msgSvc != nil {
				return h.msgSvc.OnLocationMessage(msg)
			}
		case message.MsgTypeLink:
			h.metrics.IncMessageReceived("link")
			if h.msgSvc != nil {
				return h.msgSvc.OnLinkMessage(msg)
			}
		case message.MsgTypeEvent:
			// Handle events
			switch msg.Event {
			case message.EventSubscribe:
				h.metrics.IncEventReceived("subscribe")
				if h.eventSvc != nil {
					return h.eventSvc.OnSubscribe(msg)
				}
			case message.EventUnsubscribe:
				h.metrics.IncEventReceived("unsubscribe")
				if h.eventSvc != nil {
					h.eventSvc.OnUnsubscribe(msg)
				}
			case message.EventClick:
				h.metrics.IncEventReceived("click")
				if h.eventSvc != nil {
					return h.eventSvc.OnClick(msg)
				}
			case message.EventView:
				h.metrics.IncEventReceived("view")
				if h.eventSvc != nil {
					h.eventSvc.OnView(msg)
				}
			case message.EventScan:
				h.metrics.IncEventReceived("scan")
				if h.eventSvc != nil {
					return h.eventSvc.OnScan(msg)
				}
			case message.EventLocation:
				h.metrics.IncEventReceived("location")
				if h.eventSvc != nil {
					h.eventSvc.OnLocation(msg)
				}
			}
		}
		return nil
	})

	// Serve the request (handles verification and message processing)
	if err := srv.Serve(); err != nil {
		h.log.Error("Server error", "error", err)
		h.metrics.IncMessageError("server_error")
	}
}
