package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wechat-service/internal/model"
)

// UserSQL contains all SQL statements for user operations
type UserSQL struct {
	Insert          string
	Upsert          string
	Update          string
	UpdateStatus    string
	UpdateLocation  string
	Delete          string
	FindByOpenID    string
	FindByUnionID   string
	Exists          string
	List            string
	ListByTag       string
	Count           string
	CountSubscribed string
}

// NewUserSQL creates UserSQL with all statements
func NewUserSQL(tablePrefix string) *UserSQL {
	t := tablePrefix + "users"
	return &UserSQL{
		Insert: fmt.Sprintf(`
			INSERT INTO %s (
				openid, unionid, nickname, sex, city, province, country,
				language, head_img_url, remark, group_id, subscribe_time,
				subscribe_status, latitude, longitude, precision, tags, raw_data
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
			)
		`, t),

		Upsert: fmt.Sprintf(`
			INSERT INTO %s (
				openid, unionid, nickname, sex, city, province, country,
				language, head_img_url, remark, group_id, subscribe_time,
				subscribe_status, latitude, longitude, precision, tags, raw_data,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
				CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
			) ON CONFLICT (openid) DO UPDATE SET
				unionid = EXCLUDED.unionid,
				nickname = EXCLUDED.nickname,
				sex = EXCLUDED.sex,
				city = EXCLUDED.city,
				province = EXCLUDED.province,
				country = EXCLUDED.country,
				language = EXCLUDED.language,
				head_img_url = EXCLUDED.head_img_url,
				remark = EXCLUDED.remark,
				group_id = EXCLUDED.group_id,
				subscribe_time = EXCLUDED.subscribe_time,
				subscribe_status = EXCLUDED.subscribe_status,
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				precision = EXCLUDED.precision,
				tags = EXCLUDED.tags,
				raw_data = EXCLUDED.raw_data,
				updated_at = CURRENT_TIMESTAMP
		`, t),

		Update: fmt.Sprintf(`
			UPDATE %s SET
				nickname = $2, sex = $3, city = $4, province = $5, country = $6,
				language = $7, head_img_url = $8, remark = $9, group_id = $10,
				tags = $11, raw_data = $12, updated_at = CURRENT_TIMESTAMP
			WHERE openid = $1
		`, t),

		UpdateStatus: fmt.Sprintf(`
			UPDATE %s SET
				subscribe_status = $2,
				subscribe_time = CASE WHEN $2 = 1 THEN CURRENT_TIMESTAMP ELSE subscribe_time END,
				unsubscribe_time = CASE WHEN $2 = 0 THEN CURRENT_TIMESTAMP ELSE unsubscribe_time END,
				updated_at = CURRENT_TIMESTAMP
			WHERE openid = $1
		`, t),

		UpdateLocation: fmt.Sprintf(`
			UPDATE %s SET
				latitude = $2, longitude = $3, precision = $4, updated_at = CURRENT_TIMESTAMP
			WHERE openid = $1
		`, t),

		Delete: fmt.Sprintf(`DELETE FROM %s WHERE openid = $1`, t),

		FindByOpenID: fmt.Sprintf(`SELECT * FROM %s WHERE openid = $1`, t),

		FindByUnionID: fmt.Sprintf(`SELECT * FROM %s WHERE unionid = $1`, t),

		Exists: fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE openid = $1)`, t),

		List: fmt.Sprintf(`
			SELECT * FROM %s
			WHERE ($2::VARCHAR IS NULL OR openid = $2)
			AND ($3::SMALLINT IS NULL OR subscribe_status = $3)
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5
		`, t),

		ListByTag: fmt.Sprintf(`
			SELECT * FROM %s
			WHERE tags @> $1::JSONB
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, t),

		Count: fmt.Sprintf(`SELECT COUNT(*) FROM %s`, t),

		CountSubscribed: fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE subscribe_status = 1`, t),
	}
}

// UserSQLExecutor executes user SQL operations
type UserSQLExecutor struct {
	db     *sql.DB
	sql    *UserSQL
}

// NewUserSQLExecutor creates a new UserSQLExecutor
func NewUserSQLExecutor(db *sql.DB, tablePrefix string) *UserSQLExecutor {
	return &UserSQLExecutor{
		db:  db,
		sql: NewUserSQL(tablePrefix),
	}
}

// Insert inserts a new user
func (e *UserSQLExecutor) Insert(ctx context.Context, user *model.User) error {
	_, err := e.db.ExecContext(ctx, e.sql.Insert,
		user.OpenID, user.UnionID, user.Nickname, user.Sex,
		user.City, user.Province, user.Country, user.Language,
		user.HeadImgURL, user.Remark, user.GroupID, user.SubscribeTime,
		user.SubscribeStatus, user.Latitude, user.Longitude, user.Precision,
		user.Tags, user.RawData,
	)
	return err
}

// Upsert inserts or updates a user
func (e *UserSQLExecutor) Upsert(ctx context.Context, user *model.User) error {
	_, err := e.db.ExecContext(ctx, e.sql.Upsert,
		user.OpenID, user.UnionID, user.Nickname, user.Sex,
		user.City, user.Province, user.Country, user.Language,
		user.HeadImgURL, user.Remark, user.GroupID, user.SubscribeTime,
		user.SubscribeStatus, user.Latitude, user.Longitude, user.Precision,
		user.Tags, user.RawData,
	)
	return err
}

// Update updates a user
func (e *UserSQLExecutor) Update(ctx context.Context, user *model.User) error {
	result, err := e.db.ExecContext(ctx, e.sql.Update,
		user.OpenID, user.Nickname, user.Sex, user.City, user.Province,
		user.Country, user.Language, user.HeadImgURL, user.Remark,
		user.GroupID, user.Tags, user.RawData,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateStatus updates user's subscription status
func (e *UserSQLExecutor) UpdateStatus(ctx context.Context, openid string, subscribed bool) error {
	status := 0
	if subscribed {
		status = 1
	}
	result, err := e.db.ExecContext(ctx, e.sql.UpdateStatus, openid, status)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateLocation updates user's location
func (e *UserSQLExecutor) UpdateLocation(ctx context.Context, openid string, lat, lng, precision float64) error {
	result, err := e.db.ExecContext(ctx, e.sql.UpdateLocation, openid, lat, lng, precision)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete deletes a user
func (e *UserSQLExecutor) Delete(ctx context.Context, openid string) error {
	result, err := e.db.ExecContext(ctx, e.sql.Delete, openid)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// FindByOpenID finds a user by OpenID
func (e *UserSQLExecutor) FindByOpenID(ctx context.Context, openid string) (*model.User, error) {
	user := &model.User{}
	err := e.db.QueryRowContext(ctx, e.sql.FindByOpenID, openid).Scan(
		&user.ID, &user.OpenID, &user.UnionID, &user.Nickname, &user.Sex,
		&user.City, &user.Province, &user.Country, &user.Language,
		&user.HeadImgURL, &user.Remark, &user.GroupID, &user.SubscribeTime,
		&user.UnsubscribeTime, &user.SubscribeStatus, &user.Latitude,
		&user.Longitude, &user.Precision, &user.Tags, &user.RawData,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Exists checks if a user exists
func (e *UserSQLExecutor) Exists(ctx context.Context, openid string) (bool, error) {
	var exists bool
	err := e.db.QueryRowContext(ctx, e.sql.Exists, openid).Scan(&exists)
	return exists, err
}

// List retrieves users with pagination
func (e *UserSQLExecutor) List(ctx context.Context, openid string, status *int, limit, offset int) ([]*model.User, error) {
	rows, err := e.db.QueryContext(ctx, e.sql.List, openid, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID, &user.OpenID, &user.UnionID, &user.Nickname, &user.Sex,
			&user.City, &user.Province, &user.Country, &user.Language,
			&user.HeadImgURL, &user.Remark, &user.GroupID, &user.SubscribeTime,
			&user.UnsubscribeTime, &user.SubscribeStatus, &user.Latitude,
			&user.Longitude, &user.Precision, &user.Tags, &user.RawData,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// Count returns total user count
func (e *UserSQLExecutor) Count(ctx context.Context) (int, error) {
	var count int
	err := e.db.QueryRowContext(ctx, e.sql.Count).Scan(&count)
	return count, err
}

// CountSubscribed returns subscribed user count
func (e *UserSQLExecutor) CountSubscribed(ctx context.Context) (int, error) {
	var count int
	err := e.db.QueryRowContext(ctx, e.sql.CountSubscribed).Scan(&count)
	return count, err
}

// ScanUser scans a user row into model.User
func ScanUser(row *sql.Row) (*model.User, error) {
	user := &model.User{}
	err := row.Scan(
		&user.ID, &user.OpenID, &user.UnionID, &user.Nickname, &user.Sex,
		&user.City, &user.Province, &user.Country, &user.Language,
		&user.HeadImgURL, &user.Remark, &user.GroupID, &user.SubscribeTime,
		&user.UnsubscribeTime, &user.SubscribeStatus, &user.Latitude,
		&user.Longitude, &user.Precision, &user.Tags, &user.RawData,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
