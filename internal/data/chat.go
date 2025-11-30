package data

import (
	"context"
	"fmt"

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

func (mr *MessageRepo) GetMessagesList(ctx context.Context, message common.MessageRequest) ([]common.MessageResponse, error) {
	rspString, ok, err := mr.data.Cache().Get(ctx, "message_list_"+message.Uuid+"_"+message.FriendUuid)
	if !ok and err == nil {
		// 缓存没有命中
		var messageList []bizChat.MessageTB
		if res := mr.data.DB().WithContext(ctx).Table("t_message").Where("(from_uuid = ? AND to_uuid = ?) OR (from_uuid = ? AND to_uuid = ?)",
			message.Uuid, message.FriendUuid, message.FriendUuid, message.Uuid).Order("created_at ASC").Find(&messageList); res.Error != nil {
			mr.log.Errorf("GetMessagesList err: %v\n", res.Error)
			return nil, res.Error
		}
		var messageResponse []common.MessageResponse
		for _, message := range messageList {
			messageResponse = append(messageResponse, common.MessageResponse{
				ID: message.ID,
				FromUserId: message.FromUserId,
				ToUserId: message.ToUserId,
				Content: message.Content,
				ContentType: message.ContentType,
				CreatedAt: message.CreatedAt,
				FromUsername: message.FromUsername,
				ToUsername: message.ToUsername,
				Avatar: message.Avatar,
				Url: message.Url,
				FileSize: message.FileSize,
				FileName: message.FileName,
				FileType: message.FileType,
				CreatedAt: message.CreatedAt,
				MessageType: message.MessageType,
			})
		}

		return messageResponse, nil
	} else if ok && err == nil {
		mr.log.Infof("GetMessagesList cache hit: %v\n", rspString)
		var messageResponse []common.MessageResponse
		if err := json.Unmarshal([]byte(rspString), &messageResponse); err != nil {
			mr.log.Errorf("GetMessagesList unmarshal err: %v\n", err)
			return nil, err
		}
		return messageResponse, nil
	} else {
		mr.log.Errorf("GetMessagesList err: %v\n", err)
		return nil, err
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

	res := db.Limit(pageSize).Offset(int(offset)).Find(&messageList)
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
