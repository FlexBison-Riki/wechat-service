package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"

	"wechat-service/internal/config"
	"wechat-service/pkg/cache"
	"wechat-service/pkg/logger"
)

// Common errors
var (
	ErrQuotaExceeded = errors.New("API quota exceeded")
)

// Limiter manages rate limiting for API calls
type Limiter struct {
	cfg      *config.Config
	cache    cache.Cache
	log      *logger.Logger
	mu       sync.RWMutex
	counters map[string]*apiCounter
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// apiCounter tracks API usage
type apiCounter struct {
	count     int
	resetTime time.Time
	mu        sync.Mutex
}

// NewLimiter creates a new rate limiter
func NewLimiter(cfg *config.Config, cache cache.Cache, log *logger.Logger) *Limiter {
	l := &Limiter{
		cfg:      cfg,
		cache:    cache,
		log:      log,
		counters: make(map[string]*apiCounter),
		stopCh:   make(chan struct{}),
	}

	if cfg.RateLimit.Enabled {
		l.startDailyReset()
	}

	return l
}

// Allow checks if an API call is allowed
func (l *Limiter) Allow(apiName string) (bool, error) {
	if !l.cfg.RateLimit.Enabled {
		return true, nil
	}

	quota := l.getQuota(apiName)
	if quota <= 0 {
		return true, nil
	}

	count, err := l.getCount(apiName)
	if err != nil {
		return true, nil
	}

	if count >= quota {
		return false, ErrQuotaExceeded
	}

	l.increment(apiName)
	return true, nil
}

// AllowContext checks with context support
func (l *Limiter) AllowContext(ctx context.Context, apiName string) (bool, error) {
	if !l.cfg.RateLimit.Enabled {
		return true, nil
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return l.Allow(apiName)
	}
}

// getQuota returns the quota for an API
func (l *Limiter) getQuota(apiName string) int {
	if l.cfg.RateLimit.APIQuotas == nil {
		return 0
	}
	return l.cfg.RateLimit.APIQuotas[apiName]
}

// getCount gets current API usage count
func (l *Limiter) getCount(apiName string) (int, error) {
	l.mu.RLock()
	counter, ok := l.counters[apiName]
	l.mu.RUnlock()

	if !ok {
		return 0, nil
	}

	counter.mu.Lock()
	defer counter.mu.Unlock()

	if time.Now().After(counter.resetTime) {
		return 0, nil
	}

	return counter.count, nil
}

// increment increases the counter
func (l *Limiter) increment(apiName string) {
	l.mu.Lock()
	counter, ok := l.counters[apiName]
	if !ok {
		counter = &apiCounter{
			resetTime: l.getResetTime(),
		}
		l.counters[apiName] = counter
	}
	l.mu.Unlock()

	counter.mu.Lock()
	defer counter.mu.Unlock()

	if time.Now().After(counter.resetTime) {
		counter.count = 0
		counter.resetTime = l.getResetTime()
	}

	counter.count++
}

// getResetTime returns the daily reset time
func (l *Limiter) getResetTime() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
}

// startDailyReset starts background cleanup
func (l *Limiter) startDailyReset() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				l.resetExpired()
			case <-l.stopCh:
				return
			}
		}
	}()
}

// resetExpired resets expired counters
func (l *Limiter) resetExpired() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for _, counter := range l.counters {
		counter.mu.Lock()
		if now.After(counter.resetTime) {
			counter.count = 0
			counter.resetTime = l.getResetTime()
		}
		counter.mu.Unlock()
	}
}

// GetUsage returns current usage for an API
func (l *Limiter) GetUsage(apiName string) (int, int, error) {
	quota := l.getQuota(apiName)
	count, _ := l.getCount(apiName)
	return count, quota, nil
}

// Reset resets the quota for an API
func (l *Limiter) Reset(apiName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if counter, ok := l.counters[apiName]; ok {
		counter.mu.Lock()
		counter.count = 0
		counter.resetTime = l.getResetTime()
		counter.mu.Unlock()
	}

	return nil
}

// Stop stops the limiter
func (l *Limiter) Stop() {
	close(l.stopCh)
	l.wg.Wait()
}
