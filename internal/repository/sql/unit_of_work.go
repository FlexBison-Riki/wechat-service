package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// UnitOfWork represents a database transaction
type UnitOfWork struct {
	db     *sql.DB
	tx     *sql.Tx
	closed bool
}

// TxOption holds transaction options
type TxOption struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// NewUnitOfWork creates a new UnitOfWork
func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Begin starts a new transaction
func (u *UnitOfWork) Begin(ctx context.Context, opts ...TxOption) error {
	if u.tx != nil {
		return fmt.Errorf("transaction already started")
	}

	option := TxOption{Isolation: sql.LevelDefault}
	if len(opts) > 0 {
		option = opts[0]
	}

	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: option.Isolation,
		ReadOnly:  option.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	u.tx = tx
	return nil
}

// Commit commits the transaction
func (u *UnitOfWork) Commit() error {
	if u.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	if u.closed {
		return fmt.Errorf("transaction already closed")
	}
	err := u.tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	u.closed = true
	return nil
}

// Rollback rolls back the transaction
func (u *UnitOfWork) Rollback() error {
	if u.tx == nil {
		return nil
	}
	if u.closed {
		return nil
	}
	err := u.tx.Rollback()
	if err != nil && err != sql.ErrTxDone {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	u.closed = true
	return nil
}

// Tx returns the underlying sql.Tx for custom operations
func (u *UnitOfWork) Tx() *sql.Tx {
	return u.tx
}

// User returns a UserSQLExecutor with transaction support
func (u *UnitOfWork) User(tablePrefix string) *UserSQLExecutorTx {
	return &UserSQLExecutorTx{
		u:    u,
		sql:  NewUserSQL(tablePrefix),
	}
}

// Message returns a MessageSQLExecutor with transaction support
func (u *UnitOfWork) Message(tablePrefix string) *MessageSQLExecutorTx {
	return &MessageSQLExecutorTx{
		u:    u,
		sql:  NewMessageSQL(tablePrefix),
	}
}

// UserSQLExecutorTx executes user SQL within a transaction
type UserSQLExecutorTx struct {
	u   *UnitOfWork
	sql *UserSQL
}

// Insert inserts a new user within the transaction
func (e *UserSQLExecutorTx) Insert(ctx context.Context, user *model.User) error {
	_, err := e.u.tx.ExecContext(ctx, e.sql.Insert,
		user.OpenID, user.UnionID, user.Nickname, user.Sex,
		user.City, user.Province, user.Country, user.Language,
		user.HeadImgURL, user.Remark, user.GroupID, user.SubscribeTime,
		user.SubscribeStatus, user.Latitude, user.Longitude, user.Precision,
		user.Tags, user.RawData,
	)
	return err
}

// Upsert inserts or updates a user within the transaction
func (e *UserSQLExecutorTx) Upsert(ctx context.Context, user *model.User) error {
	_, err := e.u.tx.ExecContext(ctx, e.sql.Upsert,
		user.OpenID, user.UnionID, user.Nickname, user.Sex,
		user.City, user.Province, user.Country, user.Language,
		user.HeadImgURL, user.Remark, user.GroupID, user.SubscribeTime,
		user.SubscribeStatus, user.Latitude, user.Longitude, user.Precision,
		user.Tags, user.RawData,
	)
	return err
}

// Update updates a user within the transaction
func (e *UserSQLExecutorTx) Update(ctx context.Context, user *model.User) error {
	result, err := e.u.tx.ExecContext(ctx, e.sql.Update,
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

// UpdateStatus updates user's subscription status within the transaction
func (e *UserSQLExecutorTx) UpdateStatus(ctx context.Context, openid string, subscribed bool) error {
	status := 0
	if subscribed {
		status = 1
	}
	result, err := e.u.tx.ExecContext(ctx, e.sql.UpdateStatus, openid, status)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// Delete deletes a user within the transaction
func (e *UserSQLExecutorTx) Delete(ctx context.Context, openid string) error {
	result, err := e.u.tx.ExecContext(ctx, e.sql.Delete, openid)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

// FindByOpenID finds a user by OpenID within the transaction
func (e *UserSQLExecutorTx) FindByOpenID(ctx context.Context, openid string) (*model.User, error) {
	user := &model.User{}
	err := e.u.tx.QueryRowContext(ctx, e.sql.FindByOpenID, openid).Scan(
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

// MessageSQLExecutorTx executes message SQL within a transaction
type MessageSQLExecutorTx struct {
	u   *UnitOfWork
	sql *MessageSQL
}

// Insert inserts a new message within the transaction
func (e *MessageSQLExecutorTx) Insert(ctx context.Context, msg *model.Message) error {
	_, err := e.u.tx.ExecContext(ctx, e.sql.Insert,
		msg.MsgID, msg.MsgDataID, msg.Idx, msg.FromUser, msg.ToUser,
		msg.Direction, msg.MsgType, msg.Content, msg.MediaID, msg.ThumbMediaID,
		msg.Format, msg.PicURL, msg.LocationX, msg.LocationY, msg.Scale,
		msg.Label, msg.Title, msg.Description, msg.URL, msg.EventType,
		msg.EventKey, msg.Ticket, msg.MenuID, msg.RawData,
	)
	return err
}

// Upsert inserts or ignores a message within the transaction
func (e *MessageSQLExecutorTx) Upsert(ctx context.Context, msg *model.Message) error {
	_, err := e.u.tx.ExecContext(ctx, e.sql.Upsert,
		msg.MsgID, msg.MsgDataID, msg.Idx, msg.FromUser, msg.ToUser,
		msg.Direction, msg.MsgType, msg.Content, msg.MediaID, msg.ThumbMediaID,
		msg.Format, msg.PicURL, msg.LocationX, msg.LocationY, msg.Scale,
		msg.Label, msg.Title, msg.Description, msg.URL, msg.EventType,
		msg.EventKey, msg.Ticket, msg.MenuID, msg.RawData,
	)
	return err
}

// Update updates a message within the transaction
func (e *MessageSQLExecutorTx) Update(ctx context.Context, msg *model.Message) error {
	result, err := e.u.tx.ExecContext(ctx, e.sql.Update, msg.MsgID, msg.Content, msg.RawData)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}

// Delete deletes a message within the transaction
func (e *MessageSQLExecutorTx) Delete(ctx context.Context, msgID int64) error {
	result, err := e.u.tx.ExecContext(ctx, e.sql.Delete, msgID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrMessageNotFound
	}
	return nil
}

// FindByMsgID finds a message by msg_id within the transaction
func (e *MessageSQLExecutorTx) FindByMsgID(ctx context.Context, msgID int64) (*model.Message, error) {
	msg := &model.Message{}
	err := e.u.tx.QueryRowContext(ctx, e.sql.FindByMsgID, msgID).Scan(
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
