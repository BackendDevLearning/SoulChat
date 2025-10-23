package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	//v1 "kratos-realworld/api/conduit/v1"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/model"
)

type MessageRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewMessageRepo(data *model.Data, logger log.Logger) bizChat.MessageRepo {
	return &MessageRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (mr *MessageRepo) GetMessages(ctx context.Context, message common.MessageRequest) ([]common.MessageResponse, error) {
	messageList := []common.MessageResponse{}
	// 分页查询 1. 分页offset  2. 游标cursor
	offset := (message.Page - 1) * message.PageSize
	// 游标cursor
	cursor := offset + message.PageSize
	// 单聊消息
	if message.MessageType == 1 {
		rv := mr.data.DB().Table("t_message").
			Where("(from_uuid = ? AND to_uuid = ?) OR (from_uuid = ? AND to_uuid = ?)",
				message.Uuid, message.FriendUuid, message.FriendUuid, message.Uuid).
			Order("created_at ASC").
			Limit(int(cursor)).
			Offset(int(offset)).
			Find(&messageList)
		if rv != nil {
			return nil, rv.Error
		}
	}
	// 群聊消息
	if message.MessageType == 2 {
		rv := mr.data.DB().Table("t_message").
			Where("to_uuid = ?", message.FriendUuid).
			Order("created_at ASC").
			Limit(int(cursor)).
			Offset(int(offset)).
			Find(&messageList)
		if rv != nil {
			return nil, rv.Error
		}
	}

	return messageList, nil
}

func (mr *MessageRepo) FetchGroupMessage(ctx context.Context, toUuid string) ([]common.MessageResponse, error) {
	return nil, nil
}

func (mr *MessageRepo) SaveMessage(message *bizChat.MessageTB) error {
	rv := mr.data.DB().Create(message)
	if rv != nil {
		return rv.Error
	}
	return nil
}
