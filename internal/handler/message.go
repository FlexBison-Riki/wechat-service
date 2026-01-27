package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"

	"wechat-service/internal/model"
	"wechat-service/internal/service"
	"wechat-service/pkg/metrics"
)

// MessageHandler handles WeChat message and event callbacks
type MessageHandler struct {
	msgSvc  *service.MessageService
	userSvc *service.UserService
	wechat  *wechat.Wechat
	config  *config.Config
	metrics *metrics.Metrics
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(msgSvc *service.MessageService, userSvc *service.UserService, wc *wechat.Wechat, m *metrics.Metrics) *MessageHandler {
	cfg := &config.Config{
		AppID:     "",
		AppSecret: "",
		Token:     "",
	}
	return &MessageHandler{
		msgSvc:  msgSvc,
		userSvc: userSvc,
		wechat:  wc,
		config:  cfg,
		metrics: m,
	}
}

// SetConfig sets WeChat configuration
func (h *MessageHandler) SetConfig(appID, appSecret, token string) {
	h.config.AppID = appID
	h.config.AppSecret = appSecret
	h.config.Token = token
}

// VerifyServer handles WeChat server verification (GET request)
func (h *MessageHandler) VerifyServer(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if signature == "" || timestamp == "" || nonce == "" || echostr == "" {
		c.String(http.StatusBadRequest, "missing parameters")
		return
	}

	if !h.verifySignature(signature, timestamp, nonce) {
		c.String(http.StatusUnauthorized, "invalid signature")
		return
	}

	c.String(http.StatusOK, echostr)
}

// HandleMessage handles incoming WeChat messages (POST request)
func (h *MessageHandler) HandleMessage(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	if !h.verifySignature(signature, timestamp, nonce) {
		c.String(http.StatusUnauthorized, "invalid signature")
		return
	}

	oa := h.wechat.GetOfficialAccount(h.config)
	server := oa.GetServer(c.Request, c.Writer)

	server.SetMessageHandler(func(msg message.MixMessage) *message.Reply {
		return h.handleMessageInternal(msg)
	})

	if err := server.Serve(); err != nil {
		c.String(http.StatusInternalServerError, "server error")
		return
	}
	server.Send()
}

func (h *MessageHandler) handleMessageInternal(msg message.MixMessage) *message.Reply {
	// Record metrics
	if h.metrics != nil {
		h.metrics.RecordMessageReceived(string(msg.MsgType))
	}

	// Handle based on message type
	switch msg.MsgType {
	case message.MsgTypeText:
		return h.handleTextMessage(msg)
	case message.MsgTypeEvent:
		return h.handleEventMessage(msg)
	default:
		return h.handleDefaultMessage(msg)
	}
}

func (h *MessageHandler) handleTextMessage(msg message.MixMessage) *message.Reply {
	text := message.NewText(string(msg.Content))
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}

func (h *MessageHandler) handleEventMessage(msg message.MixMessage) *message.Reply {
	// Record event metrics
	if h.metrics != nil {
		h.metrics.RecordEventReceived(string(msg.Event))
	}

	switch msg.Event {
	case message.EventSubscribe:
		return h.handleSubscribeEvent(msg)
	case message.EventUnsubscribe:
		return h.handleUnsubscribeEvent(msg)
	case message.EventClick:
		return h.handleClickEvent(msg)
	default:
		text := message.NewText("Event received: " + string(msg.Event))
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
	}
}

func (h *MessageHandler) handleSubscribeEvent(msg message.MixMessage) *message.Reply {
	user := &model.User{
		OpenID:         string(msg.FromUserName),
		SubscribeTime:  time.Now(),
		SubscribeStatus: 1,
	}
	_ = h.userSvc.Subscribe(user)
	text := message.NewText("Welcome! 🎉")
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}

func (h *MessageHandler) handleUnsubscribeEvent(msg message.MixMessage) *message.Reply {
	_ = h.userSvc.Unsubscribe(string(msg.FromUserName))
	return nil
}

func (h *MessageHandler) handleClickEvent(msg message.MixMessage) *message.Reply {
	text := message.NewText("Menu clicked: " + string(msg.EventKey))
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}

func (h *MessageHandler) handleDefaultMessage(msg message.MixMessage) *message.Reply {
	text := message.NewText("Message received!")
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}

func (h *MessageHandler) verifySignature(signature, timestamp, nonce string) bool {
	arr := []string{h.config.Token, timestamp, nonce}
	sort.Strings(arr)
	str := strings.Join(arr, "")
	hash := sha1.New()
	hash.Write([]byte(str))
	expected := hex.EncodeToString(hash.Sum(nil))
	return signature == expected
}
