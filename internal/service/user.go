package service

import (
	"encoding/json"
	"fmt"
	"time"

	"wechat-service/internal/model"
	"wechat-service/pkg/cache"
	"wechat-service/pkg/logger"
)

// UserService handles user business logic
type UserService struct {
	cache  cache.Cache
	logger *logger.Logger
	prefix string
}

// NewUserService creates a new UserService
func NewUserService(cache cache.Cache) *UserService {
	return &UserService{
		cache:  cache,
		logger: logger.New(),
		prefix: "user:",
	}
}

// Save saves or updates a user
func (s *UserService) Save(user *model.User) error {
	key := s.prefix + user.OpenID
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store for 365 days
	if err := s.cache.Set(key, data, 365*24*time.Hour); err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}

	s.logger.Infof("User saved: openid=%s, subscribed=%v", user.OpenID, user.IsSubscribed())
	return nil
}

// GetByOpenID retrieves a user by OpenID
func (s *UserService) GetByOpenID(openid string) (*model.User, error) {
	key := s.prefix + openid
	data, err := s.cache.Get(key)
	if err != nil {
		return nil, fmt.Errorf("user not found: %s", openid)
	}

	var user model.User
	if err := json.Unmarshal(data.([]byte), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// Exists checks if a user exists
func (s *UserService) Exists(openid string) bool {
	key := s.prefix + openID
	_, err := s.cache.Get(key)
	return err == nil
}

// UpdateStatus updates user's subscription status
func (s *UserService) UpdateStatus(openid string, subscribed bool) error {
	user, err := s.GetByOpenID(openid)
	if err != nil {
		return err
	}

	if subscribed {
		user.SubscribeStatus = 1
		user.SubscribeTime = time.Now()
		user.UnsubscribeTime = nil
	} else {
		user.SubscribeStatus = 0
		now := time.Now()
		user.UnsubscribeTime = &now
	}
	user.UpdatedAt = time.Now()

	return s.Save(user)
}

// UpdateLocation updates user's location
func (s *UserService) UpdateLocation(openid string, lat, lng, precision float64) error {
	user, err := s.GetByOpenID(openid)
	if err != nil {
		return err
	}

	user.Latitude = &lat
	user.Longitude = &lng
	user.Precision = &precision
	user.UpdatedAt = time.Now()

	return s.Save(user)
}

// Subscribe marks a user as subscribed
func (s *UserService) Subscribe(openid string) error {
	return s.UpdateStatus(openid, true)
}

// Unsubscribe marks a user as unsubscribed
func (s *UserService) Unsubscribe(openid string) error {
	return s.UpdateStatus(openid, false)
}

// ListByTag retrieves users by tag
func (s *UserService) ListByTag(tagID int, limit, offset int) ([]*model.User, error) {
	// Placeholder - in production, query from database
	return []*model.User{}, nil
}

// Count returns total user count
func (s *UserService) Count() (int, error) {
	// Placeholder - in production, query from database
	return 0, nil
}

// CountSubscribed returns subscribed user count
func (s *UserService) CountSubscribed() (int, error) {
	// Placeholder - in production, query from database
	return 0, nil
}

// Delete removes a user (GDPR compliance for unsubscribe)
func (s *UserService) Delete(openid string) error {
	key := s.prefix + openid
	return s.cache.Set(key, nil, 0)
}
