package model

import "time"

// EventType constants
const (
	EventSubscribe      = "subscribe"
	EventUnsubscribe    = "unsubscribe"
	EventScan           = "SCAN"
	EventLocation       = "LOCATION"
	EventClick          = "CLICK"
	EventView           = "VIEW"
	EventViewMiniprogram = "view_miniprogram"
	EventScanCodePush   = "scancode_push"
	EventScanCodeWaitMsg = "scancode_waitmsg"
	EventPicSysPhoto    = "pic_sysphoto"
	EventPicPhotoOrAlbum = "pic_photo_or_album"
	EventPicWeixin      = "pic_weixin"
	EventLocationSelect = "location_select"
)

// Event represents a WeChat event
type Event struct {
	ID          int64     `json:"id" bson:"id"`
	OpenID      string    `json:"openid" bson:"openid"`
	EventType   string    `json:"event_type" bson:"event_type"`
	EventKey    string    `json:"event_key,omitempty" bson:"event_key,omitempty"`
	Ticket      string    `json:"ticket,omitempty" bson:"ticket,omitempty"`
	MenuID      string    `json:"menu_id,omitempty" bson:"menu_id,omitempty"`
	Latitude    float64   `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Precision   float64   `json:"precision,omitempty" bson:"precision,omitempty"`
	ScanType    string    `json:"scan_type,omitempty" bson:"scan_type,omitempty"`
	ScanResult  string    `json:"scan_result,omitempty" bson:"scan_result,omitempty"`
	PicCount    int       `json:"pic_count,omitempty" bson:"pic_count,omitempty"`
	PicMD5List  []string  `json:"pic_md5_list,omitempty" bson:"pic_md5_list,omitempty"`
	LocationX   float64   `json:"location_x,omitempty" bson:"location_x,omitempty"`
	LocationY   float64   `json:"location_y,omitempty" bson:"location_y,omitempty"`
	Scale       int       `json:"scale,omitempty" bson:"scale,omitempty"`
	Address     string    `json:"address,omitempty" bson:"address,omitempty"`
	POIName     string    `json:"poi_name,omitempty" bson:"poi_name,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

// IsMenuEvent returns true if event is menu-related
func (e *Event) IsMenuEvent() bool {
	switch e.EventType {
	case EventClick, EventView, EventViewMiniprogram,
		EventScanCodePush, EventScanCodeWaitMsg,
		EventPicSysPhoto, EventPicPhotoOrAlbum, EventPicWeixin,
		EventLocationSelect:
		return true
	}
	return false
}

// IsSubscriptionEvent returns true if event is subscribe/unsubscribe
func (e *Event) IsSubscriptionEvent() bool {
	return e.EventType == EventSubscribe || e.EventType == EventUnsubscribe
}

// IsLocationEvent returns true if event is location-related
func (e *Event) IsLocationEvent() bool {
	return e.EventType == EventLocation || e.EventType == EventLocationSelect
}
