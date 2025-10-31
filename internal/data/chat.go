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

// 分页查询
func (mr *MessageRepo) GetMessages(ctx context.Context, message common.MessageRequest) ([]common.MessageResponse, error) {
    var messageList []common.MessageResponse

    offset := (message.Page - 1) * message.PageSize
    pageSize := int(message.PageSize)
    if pageSize <= 0 {
        pageSize = 20 // 或其他默认值
    }
    db := mr.data.DB().WithContext(ctx).Table("t_message").Order("created_at ASC")

    if message.MessageType == 1 {
        db = db.Where("(from_uuid = ? AND to_uuid = ?) OR (from_uuid = ? AND to_uuid = ?)",
            message.Uuid, message.FriendUuid, message.FriendUuid, message.Uuid)
    } else if message.MessageType == 2 {
        db = db.Where("to_uuid = ?", message.FriendUuid)
    } else {
        return nil, fmt.Errorf("unknown message type: %d", message.MessageType)
    }

    res := db.Limit(pageSize).Offset(offset).Find(&messageList)
    if res.Error != nil {
        return nil, res.Error
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
