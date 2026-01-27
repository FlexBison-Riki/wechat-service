package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"

	"wechat-service/internal/model"
	"wechat-service/internal/service"
	"wechat-service/pkg/logger"
)

// MessageHandler handles WeChat message and event callbacks
type MessageHandler struct {
	msgSvc   *service.MessageService
	userSvc  *service.UserService
	wechat   *wechat.Wechat
	config   *config.Config
	logger   *logger.Logger
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(msgSvc *service.MessageService, userSvc *service.UserService, wc *wechat.Wechat) *MessageHandler {
	cfg := &config.Config{
		AppID:          "", // Set from environment
		AppSecret:      "",
		Token:          "",
		EncodingAESKey: "",
		Cache:          cache.NewMemory(),
	}
	return &MessageHandler{
		msgSvc:  msgSvc,
		userSvc: userSvc,
		wechat:  wc,
		config:  cfg,
		logger:  logger.New(),
	}
}

// SetConfig sets WeChat configuration
func (h *MessageHandler) SetConfig(appID, appSecret, token, aesKey string) {
	h.config.AppID = appID
	h.config.AppSecret = appSecret
	h.config.Token = token
	h.config.EncodingAESKey = aesKey
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
	msgSignature := c.Query("msg_signature")

	if !h.verifySignature(signature, timestamp, nonce) {
		c.String(http.StatusUnauthorized, "invalid signature")
		return
	}

	var rawMsg []byte
	if _, err := c.GetRawData(); err != nil {
		h.logger.Errorf("Failed to read request body: %v", err)
		c.String(http.StatusBadRequest, "invalid request")
		return
	}

	// Parse and handle message using SDK
	oa := h.wechat.GetOfficialAccount(h.config)
	server := oa.GetServer(c.Request, c.Writer)

	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		return h.handleMessageInternal(msg)
	})

	if err := server.Serve(); err != nil {
		h.logger.Errorf("Failed to serve message: %v", err)
		c.String(http.StatusInternalServerError, "server error")
		return
	}
	server.Send()
}

func (h *MessageHandler) handleMessageInternal(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("Received message: type=%s, from=%s", msg.MsgType, msg.FromUserName)

	// Convert to internal model and save
	internalMsg := h.toInternalMessage(msg)
	if err := h.msgSvc.Save(internalMsg); err != nil {
		h.logger.Errorf("Failed to save message: %v", err)
	}

	// Handle based on message type
	switch msg.MsgType {
	case message.MsgTypeText:
		return h.handleTextMessage(msg)
	case message.MsgTypeImage:
		return h.handleImageMessage(msg)
	case message.MsgTypeEvent:
		return h.handleEventMessage(msg)
	default:
		return h.handleDefaultMessage(msg)
	}
}

func (h *MessageHandler) handleTextMessage(msg *message.MixMessage) *message.Reply {
	// Echo back the text message
	text := message.NewText("Received: " + msg.Content)
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) handleImageMessage(msg *message.MixMessage) *message.Reply {
	// Handle image message
	text := message.NewText("Image received!")
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) handleEventMessage(msg *message.MixMessage) *message.Reply {
	switch msg.Event {
	case message.EventSubscribe:
		return h.handleSubscribeEvent(msg)
	case message.EventUnsubscribe:
		return h.handleUnsubscribeEvent(msg)
	case message.EventScan:
		return h.handleScanEvent(msg)
	case message.EventClick:
		return h.handleClickEvent(msg)
	case message.EventLocation:
		return h.handleLocationEvent(msg)
	default:
		text := message.NewText("Event received: " + msg.Event)
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: text,
		}
	}
}

func (h *MessageHandler) handleSubscribeEvent(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("User subscribed: %s", msg.FromUserName)
	welcomeText := "Welcome to our Service Account! 🎉"
	text := message.NewText(welcomeText)
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) handleUnsubscribeEvent(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("User unsubscribed: %s", msg.FromUserName)
	// Mark user as unsubscribed in database
	return nil
}

func (h *MessageHandler) handleScanEvent(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("User scanned: %s, key: %s", msg.FromUserName, msg.EventKey)
	text := message.NewText("Thanks for scanning!")
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) handleClickEvent(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("Menu clicked: %s, key: %s", msg.FromUserName, msg.EventKey)
	text := message.NewText("Menu clicked: " + msg.EventKey)
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) handleLocationEvent(msg *message.MixMessage) *message.Reply {
	h.logger.Infof("Location received: %s, lat: %s, lng: %s",
		msg.FromUserName, msg.Latitude, msg.Longitude)
	return nil
}

func (h *MessageHandler) handleDefaultMessage(msg *message.MixMessage) *message.Reply {
	text := message.NewText("Message received!")
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: text,
	}
}

func (h *MessageHandler) verifySignature(signature, timestamp, nonce string) bool {
	// Sort and concatenate token, timestamp, nonce
	arr := []string{h.config.Token, timestamp, nonce}
	sort.Strings(arr)

	// Join and hash
	str := strings.Join(arr, "")
	hash := sha1.New()
	hash.Write([]byte(str))
	expected := hex.EncodeToString(hash.Sum(nil))

	return signature == expected
}

func (h *MessageHandler) toInternalMessage(msg *message.MixMessage) *model.Message {
	return &model.Message{
		MsgID:     msg.MsgId,
		FromUser:  msg.FromUserName,
		ToUser:    msg.ToUserName,
		Direction: model.MsgDirectionIn,
		MsgType:   msg.MsgType,
		Content:   msg.Content,
		MediaID:   msg.MediaId,
		Format:    msg.Format,
		PicURL:    msg.PicUrl,
		EventType: msg.Event,
		EventKey:  msg.EventKey,
		CreatedAt: time.Unix(int64(msg.CreateTime), 0),
	}
}

// GetUser handles GET /api/v1/users/:openid
func (h *MessageHandler) GetUser(c *gin.Context) {
	openid := c.Param("openid")
	user, err := h.userSvc.GetByOpenID(openid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ListMessages handles GET /api/v1/messages
func (h *MessageHandler) ListMessages(c *gin.Context) {
	openid := c.Query("openid")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.msgSvc.List(openid, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
	})
}

// XMLResponse represents a WeChat XML response
type XMLResponse struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	Content      string   `xml:"Content,omitempty"`
}

// BuildTextResponse creates a text message response
func BuildTextResponse(toUser, fromUser, content string) []byte {
	resp := XMLResponse{
		ToUserName:   toUser,
		FromUserName: fromUser,
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      content,
	}
	data, _ := xml.Marshal(resp)
	return append([]byte(xml.Header), data...)
}
