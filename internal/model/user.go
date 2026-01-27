package model

import "time"

// User represents a WeChat user
type User struct {
	OpenID         string     `json:"openid" bson:"openid"`
	SubscribeTime  time.Time  `json:"subscribe_time" bson:"subscribe_time"`
	UnsubscribeTime *time.Time `json:"unsubscribe_time,omitempty" bson:"unsubscribe_time,omitempty"`
	SubscribeStatus int       `json:"subscribe_status" bson:"subscribe_status"`
	Nickname       string     `json:"nickname" bson:"nickname"`
	Sex            int        `json:"sex" bson:"sex"`
	City           string     `json:"city" bson:"city"`
	Province       string     `json:"province" bson:"province"`
	Country        string     `json:"country" bson:"country"`
	Language       string     `json:"language" bson:"language"`
	HeadImgURL     string     `json:"headimgurl" bson:"headimgurl"`
	Remark         string     `json:"remark" bson:"remark"`
	GroupID        int        `json:"groupid" bson:"groupid"`
	TagIDList      []int      `json:"tagid_list" bson:"tagid_list"`
	Latitude       *float64   `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude      *float64   `json:"longitude,omitempty" bson:"longitude,omitempty"`
	Precision      *float64   `json:"precision,omitempty" bson:"precision,omitempty"`
	CreatedAt      time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" bson:"updated_at"`
}

// IsSubscribed returns true if user is currently subscribed
func (u *User) IsSubscribed() bool {
	return u.SubscribeStatus == 1
}

// Location returns user's location as string
func (u *User) Location() string {
	if u.City == "" {
		return ""
	}
	return u.City + ", " + u.Province
}
