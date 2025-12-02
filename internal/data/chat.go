package data

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"

	//v1 "kratos-realworld/api/conduit/v1"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/common/req"
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

func ConvertToMessageListRes(messageList []bizChat.MessageTB) []res.GetMessageListRespond {
	var messageResponse []res.GetMessageListRespond
	for _, message := range messageList {
		messageResponse = append(messageResponse, res.GetMessageListRespond{
			Type:         int32(message.Type),
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
	key := "message_list_" + uuid1 + "_" + uuid2
	messageListStr, err := mr.data.Cache().LRange(ctx, key, -MessageListLength, -1)

	if err == nil && len(messageListStr) == 0 {
		// 缓存没有命中
		var messageList []bizChat.MessageTB
		query := "(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)"
		if dbRes := mr.data.DB().WithContext(ctx).Table("t_message").Where(query,
			uuid1, uuid2, uuid2, uuid1).Order("created_at ASC").Find(&messageList); dbRes.Error != nil {
			mr.log.Errorf("GetMessagesList err: %v\n", dbRes.Error)
			return nil, dbRes.Error
		}

		// 设置缓存
		err := mr.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
			// 缓存到消息列表
			for _, msg := range messageList {
				msgBytes, err := json.Marshal(msg)
				if err != nil {
					mr.log.Errorf("failed to marshal message for cache: %v", err)
					continue
				}
				if err := pipe.RPush(ctx, key, string(msgBytes)).Err(); err != nil {
					mr.log.Errorf("failed to rpush message in pipeline: %v", err)
				}
			}
			// 只保留最近100条消息
			if err := pipe.LTrim(ctx, key, -MessageListLength, -1).Err(); err != nil {
				mr.log.Errorf("failed to ltrim message list in pipeline: %v", err)
			}
			// 设置过期时间
			if err := pipe.Expire(ctx, key, MessageListCacheTTL).Err(); err != nil {	
				mr.log.Errorf("failed to set expire for message list: %v", err)
			}
			return nil
		})
		if err != nil {
			mr.log.Errorf("failed to set message list cache: %v", err)
		}
		return ConvertToMessageListRes(messageList), nil
	} else if err == nil && len(messageListStr) != 0 {
		// 缓存命中
		mr.log.Infof("GetMessagesList cache hit: %v\n", messageListStr)
		var messageList []bizChat.MessageTB
		for _, msgStr := range messageListStr {
			var msg bizChat.MessageTB
			if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
				mr.log.Errorf("GetMessagesList unmarshal err: %v", err)
				continue // 继续处理下一条消息
			}
			messageList = append(messageList, msg)
		}
		// 更新过期时间
		mr.data.Cache().Expire(ctx, key, MessageListCacheTTL)
		return ConvertToMessageListRes(messageList), nil
	} else {
		// redis 出错挂掉
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
	return ConvertToMessageListRes(messageList), total, nil
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
