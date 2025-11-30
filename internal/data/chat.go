package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	//v1 "kratos-realworld/api/conduit/v1"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/common/res"
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

func ConvertToMessageList(messageList []bizChat.MessageTB) []res.GetMessageListRespond {
	var messageResponse []res.GetMessageListRespond
	for _, message := range messageList {
		messageResponse = append(messageResponse, res.GetMessageListRespond{
			Type:         message.Type,
			Content:      message.Content,
			Url:          message.Url,
			SendId:       message.SendId,
			SendName:     message.SendName,
			SendAvatar:   message.SendAvatar,
			ReceiveAvatar: message.ReceiveAvatar,
			ReceiveId:    message.ReceiveId,
			FileSize:     message.FileSize,
			FileType:     message.FileType,
			FileName:     message.FileName,
			AVdata:       message.AVdata,
		})
	}
	return messageResponse
}

func (mr *MessageRepo) GetMessagesList(ctx context.Context, uuid1 string, uuid2 string) ([]res.GetMessageListRespond, error) {
	rspString, ok, err := mr.data.Cache().Get(ctx, "message_list_"+uuid1+"_"+uuid2)
	if !ok && err == nil {
		// 缓存没有命中
		var messageList []bizChat.MessageTB
		query := "(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)"
		if dbRes := mr.data.DB().WithContext(ctx).Table("t_message").Where(query,
			uuid1, uuid2, uuid2, uuid1).Order("created_at ASC").Find(&messageList); dbRes.Error != nil {
			mr.log.Errorf("GetMessagesList err: %v\n", dbRes.Error)
			return nil, dbRes.Error
		}
		return ConvertToMessageList(messageList), nil
	} else if ok && err == nil {
		// 缓存命中
		mr.log.Infof("GetMessagesList cache hit: %v\n", rspString)
		var messageList []res.GetMessageListRespond
		if err := json.Unmarshal([]byte(rspString), &messageList); err != nil {
			mr.log.Errorf("GetMessagesList unmarshal err: %v\n", err)
			return nil, err
		}
		return messageList, nil
	} else {
		mr.log.Errorf("GetMessagesList err: %v\n", err)
		return nil, err
	}
}

// 分页查询
func (mr *MessageRepo) GetMessages(ctx context.Context, message req.MessageRequest) ([]res.GetMessageListRespond, int64, error) {
	offset := (message.Page - 1) * message.PageSize
	pageSize := int(message.PageSize)
	if pageSize <= 0 {
		pageSize = 20 // 默认值
	}

	// 构建查询条件
	var query string
	var args []interface{}
	if message.MessageType == common.MESSAGE_TYPE_SINGLE {
		// 单聊：查询双方的消息
		query = "(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)"
		args = []interface{}{message.Uuid, message.FriendUuid, message.FriendUuid, message.Uuid}
	} else if message.MessageType == 2 {
		// 群聊：查询群消息
		query = "receive_id = ?"
		args = []interface{}{message.FriendUuid}
	} else {
		return nil, 0, fmt.Errorf("unknown message type: %d", message.MessageType)
	}

	// 查询总数
	var total int64
	dbCount := mr.data.DB().WithContext(ctx).Table("t_message").Where(query, args...)
	if err := dbCount.Count(&total).Error; err != nil {
		mr.log.Errorf("GetMessages count err: %v\n", err)
		return nil, 0, err
	}

	// 查询消息列表
	var messageList []bizChat.MessageTB
	db := mr.data.DB().WithContext(ctx).Table("t_message").Where(query, args...).Order("created_at ASC")
	if err := db.Limit(pageSize).Offset(int(offset)).Find(&messageList).Error; err != nil {
		mr.log.Errorf("GetMessages err: %v\n", err)
		return nil, 0, err
	}

	// 转换为响应格式
	return ConvertToMessageList(messageList), total, nil
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
