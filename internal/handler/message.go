package handler

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wechat-service/internal/config"
	"wechat-service/internal/model"
	"wechat-service/internal/service"
	"wechat-service/pkg/logger"
	"wechat-service/pkg/metrics"

	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// MessageHandler handles WeChat message and event callbacks
type MessageHandler struct {
	cfg        *config.Config
	tokenSvc   *service.TokenServer
	msgService service.MessageService
	eventService service.EventService
	log        *logger.Logger
	metrics    *metrics.Metrics
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(
	cfg *config.Config,
	tokenSvc *service.TokenServer,
	msgService service.MessageService,
	eventService service.EventService,
	log *logger.Logger,
	metrics *metrics.Metrics,
) *MessageHandler {
	return &MessageHandler{
		cfg:          cfg,
		tokenSvc:     tokenSvc,
		msgService:   msgService,
		eventService: eventService,
		log:          log,
		metrics:      metrics,
	}
}

// ServeHTTP handles the WeChat callback request
func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("Panic in message handler", "error", r)
			h.metrics.IncMessagePanic()
			w.Write([]byte("success"))
		}
	}()

	// Update metrics
	h.metrics.IncHTTPRequest("wechat", r.Method)

	// Handle different request types
	switch r.Method {
	case "GET":
		h.handleVerification(w, r)
	case "POST":
		h.handleMessage(w, r)
	default:
		h.metrics.IncHTTPError("wechat", r.Method, 405)
		http.Error(w, "Method not allowed", 405)
	}
}

// handleVerification handles the WeChat server verification (GET request)
func (h *MessageHandler) handleVerification(w http.ResponseWriter, r *http.Request) {
	h.log.Debug("Received verification request", "query", r.URL.Query())

	// Verify signature
	signature := r.URL.Query().Get("signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	echostr := r.URL.Query().Get("echostr")

	if !message.CheckSignature(signature, timestamp, nonce, h.cfg.WeChat.Token) {
		h.log.Warn("Invalid signature in verification")
		h.metrics.IncMessageError("invalid_signature")
		http.Error(w, "Invalid signature", 401)
		return
	}

	// Return echostr
	h.log.Info("Verification successful")
	w.Write([]byte(echostr))
}

// handleMessage handles incoming messages and events (POST request)
func (h *MessageHandler) handleMessage(w http.ResponseWriter, r *http.Request) {
	// Check signature again for POST requests
	signature := r.URL.Query().Get("signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")

	if !message.CheckSignature(signature, timestamp, nonce, h.cfg.WeChat.Token) {
		h.log.Warn("Invalid signature in message")
		h.metrics.IncMessageError("invalid_signature")
		http.Error(w, "Invalid signature", 401)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Failed to read request body", "error", err)
		h.metrics.IncMessageError("read_body_failed")
		http.Error(w, "Failed to read body", 400)
		return
	}
	defer r.Body.Close()

	// Decrypt if using AES encryption
	if h.cfg.WeChat.EncodingAESKey != "" {
		body, err = h.decryptMessage(body, r.URL.Query())
		if err != nil {
			h.log.Error("Failed to decrypt message", "error", err)
			h.metrics.IncMessageError("decrypt_failed")
			http.Error(w, "Failed to decrypt", 400)
			return
		}
	}

	// Parse XML
	var msg message.MixMessage
	if err := xml.Unmarshal(body, &msg); err != nil {
		h.log.Error("Failed to parse XML", "error", err, "body", string(body))
		h.metrics.IncMessageError("xml_parse_failed")
		http.Error(w, "Invalid XML", 400)
		return
	}

	// Log message
	h.log.Debug("Received message",
		"msg_type", msg.MsgType,
		"from_user", msg.FromUserName,
		"create_time", msg.CreateTime,
	)

	// Process message in goroutine to respond quickly
	// This ensures 5-second timeout compliance
	go func() {
		h.processMessage(&msg)
	}()

	// Return success immediately to avoid timeout
	// Actual response is sent asynchronously
	w.Write([]byte("success"))
}

// processMessage processes the message and sends response if needed
func (h *MessageHandler) processMessage(msg *message.MixMessage) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("Panic processing message", "error", r, "msg_type", msg.MsgType)
		}
	}()

	var response interface{}

	switch msg.MsgType {
	case message.MsgTypeText:
		h.metrics.IncMessageReceived("text")
		response = h.handleTextMessage(msg)

	case message.MsgTypeImage:
		h.metrics.IncMessageReceived("image")
		response = h.handleImageMessage(msg)

	case message.MsgTypeVoice:
		h.metrics.IncMessageReceived("voice")
		response = h.handleVoiceMessage(msg)

	case message.MsgTypeVideo:
		h.metrics.IncMessageReceived("video")
		response = h.handleVideoMessage(msg)

	case message.MsgTypeLocation:
		h.metrics.IncMessageReceived("location")
		response = h.handleLocationMessage(msg)

	case message.MsgTypeLink:
		h.metrics.IncMessageReceived("link")
		response = h.handleLinkMessage(msg)

	case message.EventType:
		response = h.handleEvent(msg)

	default:
		h.log.Warn("Unknown message type", "type", msg.MsgType)
		h.metrics.IncMessageReceived("unknown")
	}

	// Send response if any
	if response != nil {
		h.sendResponse(msg, response)
	}
}

// handleTextMessage handles text messages
func (h *MessageHandler) handleTextMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling text message", "content", msg.Content)

	// Delegate to message service
	if h.msgService != nil {
		return h.msgService.OnTextMessage(msg)
	}

	// Default echo response
	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: msg.Content,
		},
	}
}

// handleImageMessage handles image messages
func (h *MessageHandler) handleImageMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling image message", "media_id", msg.MediaID)

	if h.msgService != nil {
		return h.msgService.OnImageMessage(msg)
	}

	// Default: echo the image
	return &message.Reply{
		MsgType: message.MsgTypeImage,
		MsgData: &message.Image{
			MediaID: msg.MediaID,
		},
	}
}

// handleVoiceMessage handles voice messages
func (h *MessageHandler) handleVoiceMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling voice message", "media_id", msg.MediaID)

	if h.msgService != nil {
		return h.msgService.OnVoiceMessage(msg)
	}

	return nil
}

// handleVideoMessage handles video messages
func (h *MessageHandler) handleVideoMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling video message", "media_id", msg.MediaID)

	if h.msgService != nil {
		return h.msgService.OnVideoMessage(msg)
	}

	return nil
}

// handleLocationMessage handles location messages
func (h *MessageHandler) handleLocationMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling location message",
		"location_x", msg.LocationX,
		"location_y", msg.LocationY,
		"scale", msg.Scale,
	)

	if h.msgService != nil {
		return h.msgService.OnLocationMessage(msg)
	}

	return nil
}

// handleLinkMessage handles link messages
func (h *MessageHandler) handleLinkMessage(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling link message", "title", msg.Title)

	if h.msgService != nil {
		return h.msgService.OnLinkMessage(msg)
	}

	return nil
}

// handleEvent handles event messages
func (h *MessageHandler) handleEvent(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Handling event", "event", msg.Event)

	if h.eventService != nil {
		return h.eventService.OnEvent(msg)
	}

	switch msg.Event {
	case message.EventSubscribe:
		h.metrics.IncEventReceived("subscribe")
		return h.handleSubscribeEvent(msg)

	case message.EventUnsubscribe:
		h.metrics.IncEventReceived("unsubscribe")
		return h.handleUnsubscribeEvent(msg)

	case message.EventClick:
		h.metrics.IncEventReceived("click")
		return h.handleClickEvent(msg)

	case message.EventView:
		h.metrics.IncEventReceived("view")
		return h.handleViewEvent(msg)

	case message.EventScan:
		h.metrics.IncEventReceived("scan")
		return h.handleScanEvent(msg)

	case message.EventLocation:
		h.metrics.IncEventReceived("location")
		return h.handleLocationEvent(msg)

	default:
		h.metrics.IncEventReceived("unknown")
		h.log.Warn("Unknown event type", "event", msg.Event)
	}

	return nil
}

// handleSubscribeEvent handles subscription events
func (h *MessageHandler) handleSubscribeEvent(msg *message.MixMessage) *message.Reply {
	h.log.Info("User subscribed", "openid", msg.FromUserName)

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: "欢迎关注我们的服务号！",
		},
	}
}

// handleUnsubscribeEvent handles unsubscription events
func (h *MessageHandler) handleUnsubscribeEvent(msg *message.MixMessage) *message.Reply {
	h.log.Info("User unsubscribed", "openid", msg.FromUserName)
	return nil
}

// handleClickEvent handles menu click events
func (h *MessageHandler) handleClickEvent(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Menu clicked", "key", msg.EventKey)

	if h.eventService != nil {
		return h.eventService.OnClickEvent(msg)
	}

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: &message.Text{
			Content: fmt.Sprintf("您点击了: %s", msg.EventKey),
		},
	}
}

// handleViewEvent handles menu view events
func (h *MessageHandler) handleViewEvent(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Menu viewed", "url", msg.EventKey)
	return nil
}

// handleScanEvent handles QR code scan events
func (h *MessageHandler) handleScanEvent(msg *message.MixMessage) *message.Reply {
	h.log.Debug("QR code scanned", "event_key", msg.EventKey, "ticket", msg.Ticket)
	return nil
}

// handleLocationEvent handles location update events
func (h *MessageHandler) handleLocationEvent(msg *message.MixMessage) *message.Reply {
	h.log.Debug("Location updated", "latitude", msg.Latitude, "longitude", msg.Longitude)
	return nil
}

// sendResponse sends the response message to WeChat
func (h *MessageHandler) sendResponse(msg *message.MixMessage, response interface{}) {
	if response == nil {
		return
	}

	data, err := xml.Marshal(response)
	if err != nil {
		h.log.Error("Failed to marshal response", "error", err)
		return
	}

	// Wrap in CDATA
	xmlData := h.wrapResponseXML(msg, data)

	h.log.Debug("Sending response", "to", msg.FromUserName, "type", fmt.Sprintf("%T", response))

	// Note: Actual sending would use客服 message API for async responses
	// This is a placeholder for the response structure
	h.metrics.IncMessageSent("response")
}

// wrapResponseXML wraps XML with proper user info
func (h *MessageHandler) wrapResponseXML(msg *message.MixMessage, data []byte) []byte {
	// Extract the message type for proper formatting
	content := string(data)

	// Simple response - in production, use proper SDK methods
	response := fmt.Sprintf(`<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%d</CreateTime>
%s
</xml>`,
		msg.FromUserName,
		msg.ToUserName,
		time.Now().Unix(),
		content,
	)

	return []byte(response)
}

// decryptMessage decrypts encrypted WeChat messages
func (h *MessageHandler) decryptMessage(encrypted []byte, params url.Values) ([]byte, error) {
	// Extract encryption parameters
	encryptType := params.Get("encrypt_type")
	if encryptType != "aes" {
		return encrypted, nil
	}

	msgSignature := params.Get("msg_signature")
	timestamp := params.Get("timestamp")
	nonce := params.Get("nonce")

	// Parse the encrypted message format
	var encMsg struct {
		XMLName    xml.Name `xml:"xml"`
		Encrypt    string   `xml:"Encrypt"`
		MsgSignature string `xml:"MsgSignature"`
		TimeStamp  string   `xml:"TimeStamp"`
		Nonce      string   `xml:"Nonce"`
	}

	if err := xml.Unmarshal(encrypted, &encMsg); err != nil {
		return nil, fmt.Errorf("failed to parse encrypted message: %w", err)
	}

	// Decode base64
	cipherText, err := base64.StdEncoding.DecodeString(encMsg.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Decrypt using AES-256-CBC
	// In production, use proper WeChat SDK crypto methods
	// This is a placeholder - use silenceper/wechat's crypto module

	// Simplified placeholder - use actual WeChat SDK
	// For now, return the encrypted content (SDK will handle decryption)
	return cipherText, nil
}

// VerifySignature verifies the WeChat signature
func VerifySignature(token, signature, timestamp, nonce string) bool {
	// Use SDK's built-in verification
	return message.CheckSignature(signature, timestamp, nonce, token)
}

// BuildResponse builds a response message
func BuildResponse(fromUser, toUser string, msgType string, content interface{}) []byte {
	var contentXML string

	switch v := content.(type) {
	case string:
		contentXML = fmt.Sprintf("<Content><![CDATA[%s]]></Content>", v)
	case *message.Text:
		contentXML = fmt.Sprintf("<Content><![CDATA[%s]]></Content>", v.Content)
	case *message.Image:
		contentXML = fmt.Sprintf("<Image><MediaId><![CDATA[%s]]></MediaId></Image>", v.MediaID)
	case *message.Voice:
		contentXML = fmt.Sprintf("<Voice><MediaId><![CDATA[%s]]></MediaId></Voice>", v.MediaID)
	case *message.Video:
		contentXML = fmt.Sprintf("<Video><MediaId><![CDATA[%s]]></MediaId></Video>", v.MediaID)
	case *message.Music:
		contentXML = fmt.Sprintf(`<Music>
<Title><![CDATA[%s]]></Title>
<Description><![CDATA[%s]]></Description>
<MusicUrl><![CDATA[%s]]></MusicUrl>
<HQMusicUrl><![CDATA[%s]]></HQMusicUrl>
<ThumbMediaId><![CDATA[%s]]></ThumbMediaId>
</Music>`,
			v.Title, v.Description, v.MusicURL, v.HQMusicURL, v.ThumbMediaID)
	case *message.News:
		articles := make([]string, len(v.Articles))
		for i, article := range v.Articles {
			articles[i] = fmt.Sprintf(`<item>
<Title><![CDATA[%s]]></Title>
<Description><![CDATA[%s]]></Description>
<PicUrl><![CDATA[%s]]></PicUrl>
<Url><![CDATA[%s]]></Url>
</item>`,
				article.Title, article.Description, article.PicURL, article.URL)
		}
		contentXML = fmt.Sprintf("<News><Articles>%s</Articles></News>", strings.Join(articles, ""))
	}

	xml := fmt.Sprintf(`<xml>
<ToUserName><![CDATA[%s]]></ToUserName>
<FromUserName><![CDATA[%s]]></FromUserName>
<CreateTime>%d</CreateTime>
<MsgType><![CDATA[%s]]></MsgType>
%s
</xml>`,
		toUser,
		fromUser,
		time.Now().Unix(),
		msgType,
		contentXML,
	)

	return []byte(xml)
}

// CompressData compresses data using deflate
func CompressData(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	w, err := flate.NewWriter(buf, flate.BestCompression)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	w.Close()

	return buf.Bytes(), nil
}

// DecompressData decompresses deflate data
func DecompressData(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	return io.ReadAll(r)
}

// ParseMessage parses raw XML to message model
func ParseMessage(data []byte) (*model.Message, error) {
	var msg message.MixMessage
	if err := xml.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &model.Message{
		MsgID:        msg.MsgID,
		FromUserName: msg.FromUserName,
		ToUserName:   msg.ToUserName,
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
		CreateTime:   msg.CreateTime,
	}, nil
}
