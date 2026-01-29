package repository

import (
	"sync"
	"time"
)

// Message represents a WeChat message
type Message struct {
	ID           int64     `json:"id"`
	MsgID        int64     `json:"msg_id"`
	FromUser     string    `json:"from_user"`
	ToUser       string    `json:"to_user"`
	MsgType      string    `json:"msg_type"`
	Content      string    `json:"content"`
	MediaID      string    `json:"media_id"`
	PicURL       string    `json:"pic_url"`
	Format       string    `json:"format"`
	ThumbMediaID string    `json:"thumb_media_id"`
	LocationX    float64   `json:"location_x"`
	LocationY    float64   `json:"location_y"`
	Scale        int       `json:"scale"`
	Label        string    `json:"label"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	URL          string    `json:"url"`
	Event        string    `json:"event"`
	EventKey     string    `json:"event_key"`
	Ticket       string    `json:"ticket"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Precision    float64   `json:"precision"`
	MsgDataID    int       `json:"msg_data_id"`
	Idx          int       `json:"idx"`
	CreatedAt    time.Time `json:"created_at"`
}

// MessageRepository handles message data storage
type MessageRepository struct {
	mu       sync.RWMutex
	messages map[int64]*Message
}

// NewMessageRepository creates a new message repository
func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		messages: make(map[int64]*Message),
	}
}

// GetByMsgID retrieves a message by msg_id
func (r *MessageRepository) GetByMsgID(msgID int64) (*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, ok := r.messages[msgID]
	if !ok {
		return nil, nil
	}
	return msg, nil
}

// Create creates a new message
func (r *MessageRepository) Create(msg *Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg.CreatedAt = time.Now()
	r.messages[msg.MsgID] = msg

	return nil
}

// Save creates or updates a message
func (r *MessageRepository) Save(msg *Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	msg.CreatedAt = time.Now()
	r.messages[msg.MsgID] = msg

	return nil
}

// Delete deletes a message by msg_id
func (r *MessageRepository) Delete(msgID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.messages, msgID)
	return nil
}

// GetAll retrieves all messages
func (r *MessageRepository) GetAll() ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := make([]*Message, 0, len(r.messages))
	for _, msg := range r.messages {
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// GetByType retrieves messages by type
func (r *MessageRepository) GetByType(msgType string) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := make([]*Message, 0)
	for _, msg := range r.messages {
		if msg.MsgType == msgType {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

// GetByUser retrieves messages from a specific user
func (r *MessageRepository) GetByUser(openid string) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := make([]*Message, 0)
	for _, msg := range r.messages {
		if msg.FromUser == openid {
			msgs = append(msgs, msg)
		}
	}
	return msgs, nil
}

// GetRecent retrieves recent messages with limit
func (r *MessageRepository) GetRecent(limit int) ([]*Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := make([]*Message, 0, limit)
	for _, msg := range r.messages {
		msgs = append(msgs, msg)
	}

	// Sort by created_at desc
	for i := 0; i < len(msgs); i++ {
		for j := i + 1; j < len(msgs); j++ {
			if msgs[j].CreatedAt.After(msgs[i].CreatedAt) {
				msgs[i], msgs[j] = msgs[j], msgs[i]
			}
		}
	}

	if len(msgs) > limit {
		msgs = msgs[:limit]
	}

	return msgs, nil
}

// Count returns total message count
func (r *MessageRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.messages)
}

// CountByType returns message count by type
func (r *MessageRepository) CountByType(msgType string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, msg := range r.messages {
		if msg.MsgType == msgType {
			count++
		}
	}
	return count
}

// DeleteOld deletes messages older than duration
func (r *MessageRepository) DeleteOld(olderThan time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	for msgID, msg := range r.messages {
		if msg.CreatedAt.Before(cutoff) {
			delete(r.messages, msgID)
		}
	}

	return nil
}
