package repository

import (
	"sync"
	"time"

	"wechat-service/pkg/logger"
)

// User represents a WeChat user
type User struct {
	OpenID       string    `json:"openid"`
	UnionID      string    `json:"unionid"`
	Nickname     string    `json:"nickname"`
	Sex          int       `json:"sex"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	Country      string    `json:"country"`
	HeadImgURL   string    `json:"headimgurl"`
	Subscribe    int       `json:"subscribe"`
	SubscribeTime time.Time `json:"subscribe_time"`
	Remark       string    `json:"remark"`
	TagIDList    []int     `json:"tagid_list"`
	GroupID      int       `json:"groupid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserRepository handles user data storage
type UserRepository struct {
	mu     sync.RWMutex
	users  map[string]*User
	log    *logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[string]*User),
		log:   logger.NewLogger("info"),
	}
}

// GetByOpenID retrieves a user by openid
func (r *UserRepository) GetByOpenID(openid string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[openid]
	if !ok {
		return nil, nil
	}
	return user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[user.OpenID] = user

	r.log.Debug("User created", "openid", user.OpenID)
	return nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()
	if existing, ok := r.users[user.OpenID]; ok {
		user.CreatedAt = existing.CreatedAt
	}
	r.users[user.OpenID] = user

	r.log.Debug("User updated", "openid", user.OpenID)
	return nil
}

// Save creates or updates a user
func (r *UserRepository) Save(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()
	if existing, ok := r.users[user.OpenID]; ok {
		user.CreatedAt = existing.CreatedAt
	} else {
		user.CreatedAt = time.Now()
	}
	r.users[user.OpenID] = user

	return nil
}

// Delete deletes a user by openid
func (r *UserRepository) Delete(openid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, openid)
	r.log.Debug("User deleted", "openid", openid)
	return nil
}

// GetAll retrieves all users
func (r *UserRepository) GetAll() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

// GetByTag retrieves users by tag
func (r *UserRepository) GetByTag(tagID int) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0)
	for _, user := range r.users {
		for _, id := range user.TagIDList {
			if id == tagID {
				users = append(users, user)
				break
			}
		}
	}
	return users, nil
}

// GetSubscribed retrieves all subscribed users
func (r *UserRepository) GetSubscribed() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0)
	for _, user := range r.users {
		if user.Subscribe == 1 {
			users = append(users, user)
		}
	}
	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.users)
}

// CountSubscribed returns the number of subscribed users
func (r *UserRepository) CountSubscribed() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, user := range r.users {
		if user.Subscribe == 1 {
			count++
		}
	}
	return count
}
