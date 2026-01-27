package model

import "time"

// MessageType constants
const (
	MsgTypeText      = "text"
	MsgTypeImage     = "image"
	MsgTypeVoice     = "voice"
	MsgTypeVideo     = "video"
	MsgTypeShortVideo = "shortvideo"
	MsgTypeLocation  = "location"
	MsgTypeLink      = "link"
	MsgTypeEvent     = "event"
)

// MessageDirection constants
const (
	MsgDirectionIn  = "in"
	MsgDirectionOut = "out"
)

// Message represents a WeChat message
type Message struct {
	ID          int64     `json:"id" bson:"id"`
	MsgID       int64     `json:"msg_id" bson:"msg_id"`
	MsgDataID   string    `json:"msg_data_id,omitempty" bson:"msg_data_id,omitempty"`
	Idx         int       `json:"idx,omitempty" bson:"idx,omitempty"`
	FromUser    string    `json:"from_user" bson:"from_user"`
	ToUser      string    `json:"to_user" bson:"to_user"`
	Direction   string    `json:"direction" bson:"direction"`
	MsgType     string    `json:"msg_type" bson:"msg_type"`
	Content     string    `json:"content,omitempty" bson:"content,omitempty"`
	MediaID     string    `json:"media_id,omitempty" bson:"media_id,omitempty"`
	ThumbMediaID string   `json:"thumb_media_id,omitempty" bson:"thumb_media_id,omitempty"`
	Format      string    `json:"format,omitempty" bson:"format,omitempty"`
	PicURL      string    `json:"pic_url,omitempty" bson:"pic_url,omitempty"`
	LocationX   float64   `json:"location_x,omitempty" bson:"location_x,omitempty"`
	LocationY   float64   `json:"location_y,omitempty" bson:"location_y,omitempty"`
	Scale       int       `json:"scale,omitempty" bson:"scale,omitempty"`
	Label       string    `json:"label,omitempty" bson:"label,omitempty"`
	Title       string    `json:"title,omitempty" bson:"title,omitempty"`
	Description string    `json:"description,omitempty" bson:"description,omitempty"`
	URL         string    `json:"url,omitempty" bson:"url,omitempty"`
	EventType   string    `json:"event_type,omitempty" bson:"event_type,omitempty"`
	EventKey    string    `json:"event_key,omitempty" bson:"event_key,omitempty"`
	Ticket      string    `json:"ticket,omitempty" bson:"ticket,omitempty"`
	MenuID      string    `json:"menu_id,omitempty" bson:"menu_id,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

// IsIncoming returns true if message is from user to service
func (m *Message) IsIncoming() bool {
	return m.Direction == MsgDirectionIn
}

// IsEvent returns true if message is an event
func (m *Message) IsEvent() bool {
	return m.MsgType == MsgTypeEvent
}
