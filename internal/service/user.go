package service

import (
	"context"

	"wechat-service/internal/model"
	"wechat-service/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Save saves or updates a user
func (s *UserService) Save(user *model.User) error {
	return s.repo.Save(context.Background(), user)
}

// GetByOpenID retrieves a user by OpenID
func (s *UserService) GetByOpenID(openid string) (*model.User, error) {
	return s.repo.GetByOpenID(context.Background(), openid)
}

// Exists checks if a user exists
func (s *UserService) Exists(openid string) bool {
	return s.repo.Exists(context.Background(), openid)
}

// Subscribe handles user subscription
func (s *UserService) Subscribe(user *model.User) error {
	return s.repo.Subscribe(context.Background(), user)
}

// Unsubscribe handles user unsubscription
func (s *UserService) Unsubscribe(openid string) error {
	return s.repo.Unsubscribe(context.Background(), openid)
}

// UpdateStatus updates user's subscription status
func (s *UserService) UpdateStatus(openid string, subscribed bool) error {
	return s.repo.UpdateStatus(context.Background(), openid, subscribed)
}

// Delete removes a user
func (s *UserService) Delete(openid string) error {
	return s.repo.Delete(context.Background(), openid)
}

// List retrieves users with pagination
func (s *UserService) List(limit, offset int) ([]*model.User, error) {
	return s.repo.List(context.Background(), limit, offset)
}

// Count returns total user count
func (s *UserService) Count() int {
	return s.repo.Count(context.Background())
}
