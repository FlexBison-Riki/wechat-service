package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"wechat-service/internal/config"
	"wechat-service/pkg/cache"
	"wechat-service/pkg/logger"
)

// Common error codes
var (
	ErrTokenExpired     = errors.New("access_token expired")
	ErrTokenInvalid     = errors.New("invalid access_token")
	ErrRefreshFailed    = errors.New("failed to refresh token")
	ErrCacheUnavailable = errors.New("cache unavailable")
)

// TokenResponse represents the API response for token
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"-"` // Calculated expiry timestamp
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

// IsValid checks if the response indicates success
func (r *TokenResponse) IsValid() bool {
	return r.ErrCode == 0 && r.AccessToken != ""
}

// Server manages AccessToken with centralized control
type Server struct {
	cfg        *config.Config
	cache      cache.Cache
	log        *logger.Logger
	mu         sync.RWMutex
	token      string
	expiresAt  time.Time
	refreshAt  time.Time
	refreshing bool
	cond       *sync.Cond
	stopCh     chan struct{}
	wg         sync.WaitGroup

	// Metrics
	refreshCount    int64
	failureCount    int64
	lastRefreshTime time.Time
}

// NewServer creates a new AccessToken server
func NewServer(cfg *config.Config, cache cache.Cache, log *logger.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		cache:  cache,
		log:    log,
		cond:   sync.NewCond(&sync.Mutex{}),
		stopCh: make(chan struct{}),
	}

	// Try to load existing token from cache
	s.loadFromCache()

	// Start proactive refresh if enabled
	if cfg.AccessToken.EnableProactive {
		s.startProactiveRefresh()
	}

	return s
}

// GetToken returns a valid access token
func (s *Server) GetToken() (string, error) {
	s.mu.RLock()
	if s.isValid() {
		token := s.token
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	// Need to refresh
	return s.refreshToken()
}

// GetTokenContext returns token with context support for cancellation
func (s *Server) GetTokenContext(ctx context.Context) (string, error) {
	s.mu.RLock()
	if s.isValid() {
		token := s.token
		s.mu.RUnlock()
		return token, nil
	}
	s.mu.RUnlock()

	// Try to refresh with context support
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return s.refreshTokenWithContext(ctx)
	}
}

// isValid checks if current token is still valid
func (s *Server) isValid() bool {
	return s.token != "" && time.Now().Before(s.refreshAt)
}

// refreshToken refreshes the access token
func (s *Server) refreshToken() (string, error) {
	return s.refreshTokenWithContext(context.Background())
}

// refreshTokenWithContext refreshes token with context support
func (s *Server) refreshTokenWithContext(ctx context.Context) (string, error) {
	// Use mutex to prevent concurrent refresh
	s.cond.L.Lock()
	for s.refreshing {
		// Wait with context support
		done := make(chan struct{})
		go func() {
			s.cond.Wait()
			close(done)
		}()

		s.cond.L.Unlock()
		select {
		case <-done:
			// Refresh completed
		case <-ctx.Done():
			return "", ctx.Err()
		case <-s.stopCh:
			return "", errors.New("server stopped")
		}
		s.cond.L.Lock()

		// Check if refresh succeeded
		if s.isValid() {
			token := s.token
			s.cond.L.Unlock()
			return token, nil
		}
	}

	s.refreshing = true
	s.cond.L.Unlock()

	// Perform actual refresh
	token, err := s.doRefresh()

	s.cond.L.Lock()
	s.refreshing = false
	s.cond.Broadcast()
	s.cond.L.Unlock()

	return token, err
}

// doRefresh performs the actual token refresh
func (s *Server) doRefresh() (string, error) {
	s.log.Info("Refreshing access token...")

	var apiURL string
	if s.cfg.AccessToken.UseStableAPI {
		apiURL = fmt.Sprintf("https://%s/cgi-bin/stable/access_token", s.cfg.GetAPIEndpoint())
	} else {
		apiURL = fmt.Sprintf("https://%s/cgi-bin/token", s.cfg.GetAPIEndpoint())
	}

	params := url.Values{}
	if s.cfg.AccessToken.UseStableAPI {
		params.Set("grant_type", "client_credential")
		params.Set("appid", s.cfg.WeChat.AppID)
		params.Set("secret", s.cfg.WeChat.AppSecret)
	} else {
		params.Set("grant_type", "client_credential")
		params.Set("appid", s.cfg.WeChat.AppID)
		params.Set("secret", s.cfg.WeChat.AppSecret)
	}

	req, err := http.NewRequest("POST", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		s.recordFailure(err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.recordFailure(err)
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.recordFailure(err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		s.recordFailure(err)
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !tokenResp.IsValid() {
		s.recordFailure(errors.New(tokenResp.ErrMsg))
		s.log.Error("Failed to refresh token", "errcode", tokenResp.ErrCode, "errmsg", tokenResp.ErrMsg)
		return "", fmt.Errorf("API error %d: %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// Calculate expiry
	tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()

	// Update token
	s.mu.Lock()
	s.token = tokenResp.AccessToken
	s.expiresAt = time.Unix(tokenResp.ExpiresAt, 0)
	s.refreshAt = s.expiresAt.Add(-time.Duration(s.cfg.AccessToken.CacheDuration) * time.Second)
	s.mu.Unlock()

	// Save to cache
	s.saveToCache()

	// Update metrics
	s.mu.Lock()
	s.refreshCount++
	s.lastRefreshTime = time.Now()
	s.mu.Unlock()

	s.log.Info("Access token refreshed successfully",
		"expires_in", tokenResp.ExpiresIn,
		"refresh_count", s.refreshCount)

	return s.token, nil
}

// loadFromCache attempts to load token from cache
func (s *Server) loadFromCache() {
	if s.cache == nil {
		return
	}

	data, err := s.cache.Get("access_token")
	if err != nil {
		s.log.Warn("Failed to load token from cache", "error", err)
		return
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(data, &tokenResp); err != nil {
		s.log.Warn("Failed to parse cached token", "error", err)
		return
	}

	s.mu.Lock()
	s.token = tokenResp.AccessToken
	s.expiresAt = time.Unix(tokenResp.ExpiresAt, 0)
	s.refreshAt = s.expiresAt.Add(-time.Duration(s.cfg.AccessToken.CacheDuration) * time.Second)
	s.mu.Unlock()

	s.log.Info("Loaded token from cache", "expires_at", s.expiresAt)
}

// saveToCache saves current token to cache
func (s *Server) saveToCache() {
	if s.cache == nil {
		return
	}

	tokenResp := TokenResponse{
		AccessToken: s.token,
		ExpiresIn:   int(time.Until(s.expiresAt).Seconds()),
		ExpiresAt:   s.expiresAt.Unix(),
	}

	data, err := json.Marshal(tokenResp)
	if err != nil {
		s.log.Warn("Failed to marshal token for cache", "error", err)
		return
	}

	// Cache with expiry buffer (5 minutes before actual expiry)
	cacheDuration := time.Until(s.expiresAt) - 5*time.Minute
	if cacheDuration <= 0 {
		cacheDuration = 5 * time.Minute
	}

	if err := s.cache.Set("access_token", data, cacheDuration); err != nil {
		s.log.Warn("Failed to save token to cache", "error", err)
	}
}

// startProactiveRefresh starts background token refresh
func (s *Server) startProactiveRefresh() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(time.Duration(s.cfg.AccessToken.RefreshInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.mu.RLock()
				needsRefresh := !s.isValid()
				s.mu.RUnlock()

				if needsRefresh {
					if _, err := s.refreshToken(); err != nil {
						s.log.Error("Proactive refresh failed", "error", err)
					}
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

// recordFailure records a refresh failure for metrics
func (s *Server) recordFailure(err error) {
	s.mu.Lock()
	s.failureCount++
	s.mu.Unlock()
	s.log.Error("Token refresh failed", "error", err)
}

// Stop stops the token server
func (s *Server) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.log.Info("Token server stopped", "refresh_count", s.refreshCount, "failure_count", s.failureCount)
}

// GetStats returns current server statistics
func (s *Server) GetStats() TokenServerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return TokenServerStats{
		Token:           s.token[:min(10, len(s.token))] + "...",
		ExpiresAt:       s.expiresAt,
		RefreshAt:       s.refreshAt,
		RefreshCount:    s.refreshCount,
		FailureCount:    s.failureCount,
		LastRefreshTime: s.lastRefreshTime,
		IsValid:         s.isValid(),
	}
}

// TokenServerStats represents server statistics
type TokenServerStats struct {
	Token           string    `json:"token_preview"`
	ExpiresAt       time.Time `json:"expires_at"`
	RefreshAt       time.Time `json:"refresh_at"`
	RefreshCount    int64     `json:"refresh_count"`
	FailureCount    int64     `json:"failure_count"`
	LastRefreshTime time.Time `json:"last_refresh_time"`
	IsValid         bool      `json:"is_valid"`
}

// Helper function for Go 1.21+ compatibility
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
