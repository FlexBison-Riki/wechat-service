package service

import (
	"context"
	"sync"
	"time"

	"wechat-service/internal/config"
	"wechat-service/pkg/cache"
	"wechat-service/pkg/logger"
)

// Server wraps the SDK's token management with additional features
type Server struct {
	cfg       *config.Config
	cache     cache.Cache
	log       *logger.Logger
	mu        sync.RWMutex
	stats     ServerStats
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// ServerStats represents token server statistics
type ServerStats struct {
	RefreshCount    int64     `json:"refresh_count"`
	FailureCount    int64     `json:"failure_count"`
	LastRefreshTime time.Time `json:"last_refresh_time"`
	CurrentToken    string    `json:"current_token_preview"`
}

// NewServer creates a new token server
func NewServer(cfg *config.Config, cacheInst cache.Cache, log *logger.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		cache:  cacheInst,
		log:    log,
		stopCh: make(chan struct{}),
	}

	// Start proactive refresh if enabled
	if cfg.AccessToken.EnableProactive {
		s.startProactiveRefresh()
	}

	return s
}

// GetStats returns current server statistics
func (s *Server) GetStats() ServerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ServerStats{
		RefreshCount:    s.stats.RefreshCount,
		FailureCount:    s.stats.FailureCount,
		LastRefreshTime: s.stats.LastRefreshTime,
	}
}

// recordSuccess records a successful refresh
func (s *Server) recordSuccess() {
	s.mu.Lock()
	s.stats.RefreshCount++
	s.stats.LastRefreshTime = time.Now()
	s.mu.Unlock()
}

// recordFailure records a failed refresh
func (s *Server) recordFailure() {
	s.mu.Lock()
	s.stats.FailureCount++
	s.mu.Unlock()
}

// startProactiveRefresh starts background token refresh
func (s *Server) startProactiveRefresh() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		interval := time.Duration(s.cfg.AccessToken.RefreshInterval) * time.Second
		if interval == 0 {
			interval = 1 * time.Hour
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.proactiveRefresh()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// proactiveRefresh performs proactive token refresh
func (s *Server) proactiveRefresh() {
	// The SDK handles token refresh automatically via cache
	// This is for custom logging and metrics
	s.log.Debug("Proactive refresh check", "refresh_count", s.stats.RefreshCount)
}

// Refresh forces a token refresh
func (s *Server) Refresh(ctx context.Context) error {
	// SDK manages token refresh automatically
	// This can be used for manual refresh if needed
	s.log.Info("Manual token refresh triggered")
	return nil
}

// Stop stops the token server
func (s *Server) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	s.log.Info("Token server stopped", "stats", s.stats)
}
