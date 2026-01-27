package repository

import (
	"context"

	"wechat-service/internal/model"
)

// MessageRepository interface defines message data operations
type MessageRepository interface {
	Save(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, id int64) (*model.Message, error)
	GetByMsgID(ctx context.Context, msgID int64) (*model.Message, error)
	Exists(ctx context.Context, msgID int64) (bool, error)
	ListByUser(ctx context.Context, openid string, limit, offset int) ([]*model.Message, error)
	List(ctx context.Context, fromUser, msgType *string, startTime, endTime *time.Time, limit, offset int) ([]*model.Message, error)
	Count(ctx context.Context) (int, error)
	CountByUser(ctx context.Context, openid string) (int, error)
	Delete(ctx context.Context, id int64) error
}

// Compile-time check that DBMsgRepository implements MessageRepository
var _ MessageRepository = (*DBMsgRepository)(nil)
