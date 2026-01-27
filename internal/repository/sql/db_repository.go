package repository

import (
	"context"
	"time"

	"wechat-service/internal/model"
	"wechat-service/pkg/cache"
)

// DBUserRepository combines SQL and cache for user operations
type DBUserRepository struct {
	db    *Database
	cache cache.Cache
	ttl   time.Duration
}

// NewDBUserRepository creates a new DBUserRepository
func NewDBUserRepository(db *Database, cache cache.Cache) *DBUserRepository {
	return &DBUserRepository{
		db:    db,
		cache: cache,
		ttl:   1 * time.Hour,
	}
}

// Save saves a user with cache invalidation
func (r *DBUserRepository) Save(ctx context.Context, user *model.User) error {
	// Save to database
	if err := r.db.User().Upsert(ctx, user); err != nil {
		return err
	}

	// Invalidate cache
	return r.cache.Delete("user:" + user.OpenID)
}

// GetByOpenID retrieves a user from cache first, then database
func (r *DBUserRepository) GetByOpenID(ctx context.Context, openid string) (*model.User, error) {
	// Try cache first
	if r.cache != nil {
		data, err := r.cache.Get("user:" + openid)
		if err == nil {
			if user, ok := data.(*model.User); ok {
				return user, nil
			}
		}
	}

	// Fallback to database
	user, err := r.db.User().FindByOpenID(ctx, openid)
	if err != nil {
		return nil, err
	}

	// Cache for next request
	if r.cache != nil {
		_ = r.cache.Set("user:"+openid, user, r.ttl)
	}

	return user, nil
}

// Exists checks if a user exists (cache + database)
func (r *DBUserRepository) Exists(ctx context.Context, openid string) (bool, error) {
	// Check cache first
	if r.cache != nil {
		_, err := r.cache.Get("user:" + openid)
		if err == nil {
			return true, nil
		}
	}

	// Check database
	return r.db.User().Exists(ctx, openid)
}

// UpdateStatus updates user's subscription status
func (r *DBUserRepository) UpdateStatus(ctx context.Context, openid string, subscribed bool) error {
	// Update database
	if err := r.db.User().UpdateStatus(ctx, openid, subscribed); err != nil {
		return err
	}

	// Invalidate cache
	if r.cache != nil {
		return r.cache.Delete("user:" + openid)
	}
	return nil
}

// UpdateLocation updates user's location
func (r *DBUserRepository) UpdateLocation(ctx context.Context, openid string, lat, lng, precision float64) error {
	if err := r.db.User().UpdateLocation(ctx, openid, lat, lng, precision); err != nil {
		return err
	}

	if r.cache != nil {
		return r.cache.Delete("user:" + openid)
	}
	return nil
}

// Subscribe marks a user as subscribed
func (r *DBUserRepository) Subscribe(ctx context.Context, user *model.User) error {
	return r.db.WithTransaction(ctx, func(uow *UnitOfWork) error {
		// Create or update user
		if err := r.db.User().Upsert(ctx, user); err != nil {
			return err
		}
		// Update stats (pseudo-code)
		// return uow.Stats().IncrementSubscribers(ctx, time.Now())
		return nil
	})
}

// Unsubscribe marks a user as unsubscribed
func (r *DBUserRepository) Unsubscribe(ctx context.Context, openid string) error {
	if err := r.db.User().UpdateStatus(ctx, openid, false); err != nil {
		return err
	}

	if r.cache != nil {
		return r.cache.Delete("user:" + openid)
	}
	return nil
}

// Delete removes a user (GDPR compliance)
func (r *DBUserRepository) Delete(ctx context.Context, openid string) error {
	if err := r.db.User().Delete(ctx, openid); err != nil {
		return err
	}

	if r.cache != nil {
		return r.cache.Delete("user:" + openid)
	}
	return nil
}

// List retrieves users with pagination
func (r *DBUserRepository) List(ctx context.Context, status *int, limit, offset int) ([]*model.User, error) {
	return r.db.User().List(ctx, "", status, limit, offset)
}

// Count returns total user count
func (r *DBUserRepository) Count(ctx context.Context) (int, error) {
	return r.db.User().Count(ctx)
}

// CountSubscribed returns subscribed user count
func (r *DBUserRepository) CountSubscribed(ctx context.Context) (int, error) {
	return r.db.User().CountSubscribed(ctx)
}

// DBMsgRepository combines SQL and cache for message operations
type DBMsgRepository struct {
	db    *Database
	cache cache.Cache
	ttl   time.Duration
}

// NewDBMsgRepository creates a new DBMsgRepository
func NewDBMsgRepository(db *Database, cache cache.Cache) *DBMsgRepository {
	return &DBMsgRepository{
		db:    db,
		cache: cache,
		ttl:   30 * time.Minute,
	}
}

// Save saves a message (with deduplication)
func (r *DBMsgRepository) Save(ctx context.Context, msg *model.Message) error {
	// Use upsert for idempotency (ignore if exists)
	return r.db.Message().Upsert(ctx, msg)
}

// GetByID retrieves a message by internal ID
func (r *DBMsgRepository) GetByID(ctx context.Context, id int64) (*model.Message, error) {
	return r.db.Message().FindByID(ctx, id)
}

// GetByMsgID retrieves a message by WeChat msg_id
func (r *DBMsgRepository) GetByMsgID(ctx context.Context, msgID int64) (*model.Message, error) {
	return r.db.Message().FindByMsgID(ctx, msgID)
}

// Exists checks if a message exists
func (r *DBMsgRepository) Exists(ctx context.Context, msgID int64) (bool, error) {
	return r.db.Message().Exists(ctx, msgID)
}

// ListByUser retrieves messages for a user
func (r *DBMsgRepository) ListByUser(ctx context.Context, openid string, limit, offset int) ([]*model.Message, error) {
	return r.db.Message().ListByUser(ctx, openid, limit, offset)
}

// List retrieves messages with filters
func (r *DBMsgRepository) List(ctx context.Context, fromUser, msgType *string, startTime, endTime *time.Time, limit, offset int) ([]*model.Message, error) {
	return r.db.Message().List(ctx, fromUser, nil, msgType, startTime, endTime, limit, offset)
}

// Count returns total message count
func (r *DBMsgRepository) Count(ctx context.Context) (int, error) {
	return r.db.Message().Count(ctx)
}

// CountByUser returns message count for a user
func (r *DBMsgRepository) CountByUser(ctx context.Context, openid string) (int, error) {
	return r.db.Message().CountByUser(ctx, openid)
}

// Delete deletes a message
func (r *DBMsgRepository) Delete(ctx context.Context, id int64) error {
	return r.db.Message().Delete(ctx, id)
}
