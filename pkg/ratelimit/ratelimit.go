package ratelimit

import (
	"context"
	"encoding/json"
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
	cfg    *config.Config
	cache  cache.Cache
	log    *logger.Logger
	mu     sync.RWMutex
	counters map[string]*apiCounter
	stopCh  chan struct{}
	wg     sync.WaitGroup
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

	// Start daily reset goroutine
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
		return true, nil // No limit for this API
	}

	// Get current count
	count, err := l.getCount(apiName)
	if err != nil {
		l.log.Warn("Failed to get rate limit count", "api", apiName, "error", err)
		// Allow the request if we can't check
		return true, nil
	}

	if count >= quota {
		l.log.Warn("Rate limit exceeded", "api", apiName, "count", count, "quota", quota)
		return false, ErrQuotaExceeded
	}

	// Increment counter
	if err := l.increment(apiName); err != nil {
		l.log.Warn("Failed to increment rate limit counter", "api", apiName, "error", err)
	}

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
	l.cfg.RateLimit.RLock()
	defer l.cfg.RateLimit.RUnlock()

	if quota, ok := l.cfg.RateLimit.APIQuotas[apiName]; ok {
		return quota
	}
	return 0
}

// getCount gets current API usage count
func (l *Limiter) getCount(apiName string) (int, error) {
	// Try Redis first if available
	if l.cache != nil && l.cfg.RateLimit.Storage == "redis" {
		return l.getCountFromCache(apiName)
	}

	// Use local memory
	return l.getCountFromMemory(apiName)
}

// getCountFromCache gets count from Redis
func (l *Limiter) getCountFromCache(apiName string) (int, error) {
	key := l.cfg.RateLimit.Prefix + "ratelimit:" + apiName
	data, err := l.cache.Get(key)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	var counter apiCounter
	if err := decodeCounter(data, &counter); err != nil {
		return 0, err
	}

	// Check if expired
	if time.Now().After(counter.resetTime) {
		return 0, nil
	}

	return counter.count, nil
}

// getCountFromMemory gets count from local memory
func (l *Limiter) getCountFromMemory(apiName string) (int, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	counter, ok := l.counters[apiName]
	if !ok {
		return 0, nil
	}

	counter.mu.Lock()
	defer counter.mu.Unlock()

	// Check if expired
	if time.Now().After(counter.resetTime) {
		return 0, nil
	}

	return counter.count, nil
}

// increment increases the counter
func (l *Limiter) increment(apiName string) error {
	// Try Redis first if available
	if l.cache != nil && l.cfg.RateLimit.Storage == "redis" {
		return l.incrementCache(apiName)
	}

	return l.incrementMemory(apiName)
}

// incrementCache increments in Redis
func (l *Limiter) incrementCache(apiName string) error {
	key := l.cfg.RateLimit.Prefix + "ratelimit:" + apiName

	// Get current state
	current, _ := l.getCountFromCache(apiName)

	counter := &apiCounter{
		count:     current + 1,
		resetTime: l.getResetTime(),
	}

	data, err := encodeCounter(counter)
	if err != nil {
		return err
	}

	// Store with TTL (24 hours + buffer)
	return l.cache.Set(key, data, 25*time.Hour)
}

// incrementMemory increments in local memory
func (l *Limiter) incrementMemory(apiName string) error {
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

	// Check if expired
	if time.Now().After(counter.resetTime) {
		counter.count = 0
		counter.resetTime = l.getResetTime()
	}

	counter.count++
	return nil
}

// getResetTime returns the daily reset time (midnight)
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
	for apiName, counter := range l.counters {
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
	count, err := l.getCount(apiName)
	if err != nil {
		return 0, quota, err
	}
	return count, quota, nil
}

// Reset resets the quota for an API
func (l *Limiter) Reset(apiName string) error {
	// Try Redis first
	if l.cache != nil && l.cfg.RateLimit.Storage == "redis" {
		key := l.cfg.RateLimit.Prefix + "ratelimit:" + apiName
		return l.cache.Delete(key)
	}

	// Reset memory
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

// encodeCounter encodes counter to bytes
func encodeCounter(c *apiCounter) ([]byte, error) {
	data := struct {
		Count     int       `json:"count"`
		ResetTime time.Time `json:"reset_time"`
	}{
		Count:     c.count,
		ResetTime: c.resetTime,
	}
	return json.Marshal(data)
}

// decodeCounter decodes counter from bytes
func decodeCounter(data []byte, c *apiCounter) error {
	var result struct {
		Count     int       `json:"count"`
		ResetTime time.Time `json:"reset_time"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	c.count = result.Count
	c.resetTime = result.ResetTime
	return nil
}
