package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"wechat-service/internal/model"
)

// MessageSQL contains all SQL statements for message operations
type MessageSQL struct {
	Insert          string
	Upsert          string
	Update          string
	Delete          string
	FindByID        string
	FindByMsgID     string
	Exists          string
	List            string
	ListByUser      string
	ListByTimeRange string
	Count           string
	CountByUser     string
	CountByType     string
}

// NewMessageSQL creates MessageSQL with all statements
func NewMessageSQL(tablePrefix string) *MessageSQL {
	t := tablePrefix + "messages"
	return &MessageSQL{
		Insert: fmt.Sprintf(`
			INSERT INTO %s (
				msg_id, msg_data_id, idx, from_user, to_user, direction,
				msg_type, content, media_id, thumb_media_id, format, pic_url,
				location_x, location_y, scale, label, title, description, url,
				event_type, event_key, ticket, menu_id, raw_data
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
				$16, $17, $18, $19, $20, $21, $22, $23, $24
			)
		`, t),

		Upsert: fmt.Sprintf(`
			INSERT INTO %s (
				msg_id, msg_data_id, idx, from_user, to_user, direction,
				msg_type, content, media_id, thumb_media_id, format, pic_url,
				location_x, location_y, scale, label, title, description, url,
				event_type, event_key, ticket, menu_id, raw_data
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
				$16, $17, $18, $19, $20, $21, $22, $23, $24
			) ON CONFLICT (msg_id) DO NOTHING
		`, t),

		Update: fmt.Sprintf(`
			UPDATE %s SET
				content = $2, raw_data = $3
			WHERE msg_id = $1
		`, t),

		Delete: fmt.Sprintf(`DELETE FROM %s WHERE msg_id = $1`, t),

		FindByID: fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, t),

		FindByMsgID: fmt.Sprintf(`SELECT * FROM %s WHERE msg_id = $1`, t),

		Exists: fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE msg_id = $1)`, t),

		List: fmt.Sprintf(`
			SELECT * FROM %s
			WHERE ($2::VARCHAR IS NULL OR from_user = $2)
			AND ($3::VARCHAR IS NULL OR to_user = $3)
			AND ($4::VARCHAR IS NULL OR msg_type = $4)
			AND ($5::TIMESTAMPTZ IS NULL OR created_at >= $5)
			AND ($6::TIMESTAMPTZ IS NULL OR created_at <= $6)
			ORDER BY created_at DESC
			LIMIT $7 OFFSET $8
		`, t),

		ListByUser: fmt.Sprintf(`
			SELECT * FROM %s
			WHERE from_user = $1 OR to_user = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, t),

		ListByTimeRange: fmt.Sprintf(`
			SELECT * FROM %s
			WHERE created_at BETWEEN $1 AND $2
			ORDER BY created_at ASC
			LIMIT $3 OFFSET $4
		`, t),

		Count: fmt.Sprintf(`SELECT COUNT(*) FROM %s`, t),

		CountByUser: fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE from_user = $1`, t),

		CountByType: fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE msg_type = $1`, t),
	}
}

// MessageSQLExecutor executes message SQL operations
type MessageSQLExecutor struct {
	db   *sql.DB
	sql  *MessageSQL
}

// NewMessageSQLExecutor creates a new MessageSQLExecutor
func NewMessageSQLExecutor(db *sql.DB, tablePrefix string) *MessageSQLExecutor {
	return &MessageSQLExecutor{
		db:  db,
		sql: NewMessageSQL(tablePrefix),
	}
}

// Insert inserts a new message
func (e *MessageSQLExecutor) Insert(ctx context.Context, msg *model.Message) error {
	_, err := e.db.ExecContext(ctx, e.sql.Insert,
		msg.MsgID, msg.MsgDataID, msg.Idx, msg.FromUser, msg.ToUser,
		msg.Direction, msg.MsgType, msg.Content, msg.MediaID, msg.ThumbMediaID,
		msg.Format, msg.PicURL, msg.LocationX, msg.LocationY, msg.Scale,
		msg.Label, msg.Title, msg.Description, msg.URL, msg.EventType,
		msg.EventKey, msg.Ticket, msg.MenuID, msg.RawData,
	)
	return err
}

// Upsert inserts or ignores if exists (for idempotency)
func (e *MessageSQLExecutor) Upsert(ctx context.Context, msg *model.Message) error {
	_, err := e.db.ExecContext(ctx, e.sql.Upsert,
		msg.MsgID, msg.MsgDataID, msg.Idx, msg.FromUser, msg.ToUser,
		msg.Direction, msg.MsgType, msg.Content, msg.MediaID, msg.ThumbMediaID,
		msg.Format, msg.PicURL, msg.LocationX, msg.LocationY, msg.Scale,
		msg.Label, msg.Title, msg.Description, msg.URL, msg.EventType,
		msg.EventKey, msg.Ticket, msg.MenuID, msg.RawData,
	)
	return err
}

// Update updates a message
func (e *MessageSQLExecutor) Update(ctx context.Context, msg *model.Message) error {
	result, err := e.db.ExecContext(ctx, e.sql.Update, msg.MsgID, msg.Content, msg.RawData)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}

// Delete deletes a message
func (e *MessageSQLExecutor) Delete(ctx context.Context, msgID int64) error {
	result, err := e.db.ExecContext(ctx, e.sql.Delete, msgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}

// FindByID finds a message by internal ID
func (e *MessageSQLExecutor) FindByID(ctx context.Context, id int64) (*model.Message, error) {
	msg := &model.Message{}
	err := e.db.QueryRowContext(ctx, e.sql.FindByID, id).Scan(
		&msg.ID, &msg.MsgID, &msg.MsgDataID, &msg.Idx, &msg.FromUser,
		&msg.ToUser, &msg.Direction, &msg.MsgType, &msg.Content, &msg.MediaID,
		&msg.ThumbMediaID, &msg.Format, &msg.PicURL, &msg.LocationX, &msg.LocationY,
		&msg.Scale, &msg.Label, &msg.Title, &msg.Description, &msg.URL,
		&msg.EventType, &msg.EventKey, &msg.Ticket, &msg.MenuID, &msg.RawData,
		&msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// FindByMsgID finds a message by WeChat msg_id
func (e *MessageSQLExecutor) FindByMsgID(ctx context.Context, msgID int64) (*model.Message, error) {
	msg := &model.Message{}
	err := e.db.QueryRowContext(ctx, e.sql.FindByMsgID, msgID).Scan(
		&msg.ID, &msg.MsgID, &msg.MsgDataID, &msg.Idx, &msg.FromUser,
		&msg.ToUser, &msg.Direction, &msg.MsgType, &msg.Content, &msg.MediaID,
		&msg.ThumbMediaID, &msg.Format, &msg.PicURL, &msg.LocationX, &msg.LocationY,
		&msg.Scale, &msg.Label, &msg.Title, &msg.Description, &msg.URL,
		&msg.EventType, &msg.EventKey, &msg.Ticket, &msg.MenuID, &msg.RawData,
		&msg.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrMessageNotFound
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// Exists checks if a message exists
func (e *MessageSQLExecutor) Exists(ctx context.Context, msgID int64) (bool, error) {
	var exists bool
	err := e.db.QueryRowContext(ctx, e.sql.Exists, msgID).Scan(&exists)
	return exists, err
}

// List retrieves messages with filters
func (e *MessageSQLExecutor) List(ctx context.Context, fromUser, toUser, msgType *string, startTime, endTime *time.Time, limit, offset int) ([]*model.Message, error) {
	rows, err := e.db.QueryContext(ctx, e.sql.List,
		fromUser, toUser, msgType, startTime, endTime, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.MsgID, &msg.MsgDataID, &msg.Idx, &msg.FromUser,
			&msg.ToUser, &msg.Direction, &msg.MsgType, &msg.Content, &msg.MediaID,
			&msg.ThumbMediaID, &msg.Format, &msg.PicURL, &msg.LocationX, &msg.LocationY,
			&msg.Scale, &msg.Label, &msg.Title, &msg.Description, &msg.URL,
			&msg.EventType, &msg.EventKey, &msg.Ticket, &msg.MenuID, &msg.RawData,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// ListByUser retrieves messages for a specific user
func (e *MessageSQLExecutor) ListByUser(ctx context.Context, openid string, limit, offset int) ([]*model.Message, error) {
	rows, err := e.db.QueryContext(ctx, e.sql.ListByUser, openid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.MsgID, &msg.MsgDataID, &msg.Idx, &msg.FromUser,
			&msg.ToUser, &msg.Direction, &msg.MsgType, &msg.Content, &msg.MediaID,
			&msg.ThumbMediaID, &msg.Format, &msg.PicURL, &msg.LocationX, &msg.LocationY,
			&msg.Scale, &msg.Label, &msg.Title, &msg.Description, &msg.URL,
			&msg.EventType, &msg.EventKey, &msg.Ticket, &msg.MenuID, &msg.RawData,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// Count returns total message count
func (e *MessageSQLExecutor) Count(ctx context.Context) (int, error) {
	var count int
	err := e.db.QueryRowContext(ctx, e.sql.Count).Scan(&count)
	return count, err
}

// CountByUser returns message count for a user
func (e *MessageSQLExecutor) CountByUser(ctx context.Context, openid string) (int, error) {
	var count int
	err := e.db.QueryRowContext(ctx, e.sql.CountByUser, openid).Scan(&count)
	return count, err
}
