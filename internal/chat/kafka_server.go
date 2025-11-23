package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kratos-realworld/internal/kafka"
	"kratos-realworld/internal/model"
	"kratos-realworld/internal/common/respond"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// 消息类型常量
const (
	MessageTypeText         = "Text"
	MessageTypeFile         = "File"
	MessageTypeAudioOrVideo = "AudioOrVideo"
)

// 消息状态常量
const (
	MessageStatusUnsent = "Unsent"
	MessageStatusSent   = "Sent"
)

// 常量定义
const (
	REDIS_TIMEOUT = 30 // Redis 缓存超时时间（分钟）
)

// AVData 音视频数据
type AVData struct {
	MessageId string `json:"messageId"`
	Type      string `json:"type"`
}

type KafkaServerUseCase struct {
	Clients map[string]*Client
	mutex   *sync.Mutex
	Login   chan *Client // 登录通道
	Logout  chan *Client // 退出登录通道
	log  	*log.Helper
	data    *model.Data
	kafkaService *kafka.KafkaService
}

// 用来接收操作系统的信号
var kafkaQuit = make(chan os.Signal, 1)

func NewKafkaServerUseCase(log *log.Helper, data *model.Data, kafkaService *kafka.KafkaService) *KafkaServerUseCase {
	return &KafkaServerUseCase{
		Clients: make(map[string]*Client),
		mutex:   &sync.Mutex{},
		Login:   make(chan *Client),
		Logout:  make(chan *Client),
		log: 	log,
		data:    data,
		kafkaService: kafkaService,
	}
}

// Start 启动 Kafka 服务器
func (k *KafkaServerUseCase) Start() {
	defer func() {
		if r := recover(); r != nil {
			if k.log != nil {
				logger.Error("kafka server panic", zap.Any("error", r))
			}
		}
		close(k.Login)
		close(k.Logout)
	}()

	// read chat message
	go func() {
		defer func() {
			// 安全捕获 panic，防止kafka server崩溃
			if r := recover(); r != nil {
				if k.log != nil {
					k.log.Error("kafka server panic in goroutine", zap.Any("error", r))
				}
			}
		}()
		for {
			if kafka.KafkaService.ChatReader == nil {
				if k.log != nil {
					logger.Warn("kafka reader is not initialized")
				}
				time.Sleep(time.Second)
				continue
			}

			kafkaMessage, err := kafka.KafkaService.ChatReader.ReadMessage(context.Background())
			if err != nil {
				if k.log != nil {
					k.log.Error("failed to read kafka message", zap.Error(err))
				}
				continue
			}

			if k.log != nil {
				k.log.Info("received kafka message",
					zap.String("topic", kafkaMessage.Topic),
					zap.Int("partition", kafkaMessage.Partition),
					zap.Int64("offset", kafkaMessage.Offset),
					zap.String("key", string(kafkaMessage.Key)),
					zap.String("value", string(kafkaMessage.Value)),
				)
			}

			data := kafkaMessage.Value
			var chatMessageReq ChatMessageRequest
			if err := json.Unmarshal(data, &chatMessageReq); err != nil {
				if k.log != nil {
					k.log.Error("failed to unmarshal chat message", zap.Error(err))
				}
				continue
			}

			k.log.Info("原消息为：", data, "反序列化后为：", chatMessageReq)

			if chatMessageReq.Type == MessageTypeText {
				// 存message
				message := model.Message{
					Uuid:       fmt.Sprintf("M%s", GetNowAndLenRandomString(11)),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type,
					Content:    chatMessageReq.Content,
					Url:        "",
					SendId:     chatMessageReq.SendId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: chatMessageReq.SendAvatar,
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   "0B",
					FileType:   "",
					FileName:   "",
					Status:     message_status_enum.Unsent,
					CreatedAt:  time.Now(),
					AVdata:     "",
				}
				// 对SendAvatar去除前面/static之前的所有内容，防止ip前缀引入
				message.SendAvatar = normalizePath(message.SendAvatar)
				if res := dao.GormDB.Create(&message); res.Error != nil {
					zlog.Error(res.Error.Error())
				}
				if message.ReceiveId[0] == 'U' { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := respond.GetMessageListRespond{
						SendId:     message.SendId,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       message.Type,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Error("failed to marshal message", zap.Error(err))
					}
					k.log.Info("返回的消息为：", messageRsp, "序列化后为：", jsonMessage)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					k.mutex.Lock()
					if receiveClient, ok := k.Clients[message.ReceiveId]; ok {
						//messageBack.Message = jsonMessage
						//messageBack.Uuid = message.Uuid
						receiveClient.SendBack <- messageBack // 向client.Send发送
					}
					// 因为send_id肯定在线，所以这里在后端进行在线回显message，其实优化的话前端可以直接回显
					// 问题在于前后端的req和rsp结构不同，前端存储message的messageList不能存req，只能存rsp
					// 所以这里后端进行回显，前端不回显
					sendClient := k.Clients[message.SendId]
					sendClient.SendBack <- messageBack
					// 即使操作不同的 key，它们可能：
					// 共享同一个 map 结构体（头部信息）
					// 触发 map 的扩容（rehashing）
					// 修改 map 的内部计数器
					k.mutex.Unlock()

					// redis
					var rspString string
					rspString, err = k.data.Cache().Get(ctx, "message_list_" + message.SendId + "_" + message.ReceiveId).Result()
					if err == nil {
						var rsp []respond.GetMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Error("failed to unmarshal message", zap.Error(err))
						}
						rsp = append(rsp, messageRsp)
						rspByte, err := json.Marshal(rsp)
						if err != nil {
							k.log.Error("failed to marshal message", zap.Error(err))
						}
						if err := k.data.Cache().Set(ctx, "message_list_"+message.SendId+"_"+message.ReceiveId, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
							k.log.Error("failed to set message", zap.Error(err))
						}
					} else {
						if !errors.Is(err, redis.Nil) {
							k.log.Error("failed to get message", zap.Error(err))
						}
					}

				} else if message.ReceiveId[0] == 'G' { // 发送给Group
					messageRsp := respond.GetGroupMessageListRespond{
						SendId:     message.SendId,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       message.Type,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Error("failed to marshal message", zap.Error(err))
					}
					k.log.Info("返回的消息为：", messageRsp, "序列化后为：", jsonMessage)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					var group model.GroupInfo
					if res := dao.GormDB.Where("uuid = ?", message.ReceiveId).First(&group); res.Error != nil {
						k.log.Error("failed to get group", zap.Error(res.Error))
					}
					var members []string
					if err := json.Unmarshal(group.Members, &members); err != nil {
						k.log.Error("failed to unmarshal members", zap.Error(err))
					}
					k.mutex.Lock()
					for _, member := range members {
						if member != message.SendId {
							if receiveClient, ok := k.Clients[member]; ok {
								receiveClient.SendBack <- messageBack
							}
						} else {
							sendClient := k.Clients[message.SendId]
							sendClient.SendBack <- messageBack
						}
					}
					k.mutex.Unlock()

					// redis
					var rspString string
					rspString, err = myredis.GetKeyNilIsErr("group_messagelist_" + message.ReceiveId)
					if err == nil {
						var rsp []respond.GetGroupMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Error("failed to unmarshal message", zap.Error(err))
						}
						rsp = append(rsp, messageRsp)
						rspByte, err := json.Marshal(rsp)
						if err != nil {
							k.log.Error("failed to marshal message", zap.Error(err))
						}
						if err := myredis.SetKeyEx("group_messagelist_"+message.ReceiveId, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
							k.log.Error("failed to set message", zap.Error(err))
						}
					} else {
						if !errors.Is(err, redis.Nil) {
							k.log.Error("failed to get message", zap.Error(err))
						}
					}
				}
			} else if chatMessageReq.Type == message_type_enum.File {
				// 存message
				message := model.Message{
					Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type,
					Content:    "",
					Url:        chatMessageReq.Url,
					SendId:     chatMessageReq.SendId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: chatMessageReq.SendAvatar,
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   chatMessageReq.FileSize,
					FileType:   chatMessageReq.FileType,
					FileName:   chatMessageReq.FileName,
					Status:     message_status_enum.Unsent,
					CreatedAt:  time.Now(),
					AVdata:     "",
				}
				// 对SendAvatar去除前面/static之前的所有内容，防止ip前缀引入
				message.SendAvatar = normalizePath(message.SendAvatar)
				if res := dao.GormDB.Create(&message); res.Error != nil {
					zlog.Error(res.Error.Error())
				}
				if message.ReceiveId[0] == 'U' { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := respond.GetMessageListRespond{
						SendId:     message.SendId,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       message.Type,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Error("failed to marshal message", zap.Error(err))
					}
					k.log.Info("返回的消息为：", messageRsp, "序列化后为：", jsonMessage)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					k.mutex.Lock()
					if receiveClient, ok := k.Clients[message.ReceiveId]; ok {
						//messageBack.Message = jsonMessage
						//messageBack.Uuid = message.Uuid
						receiveClient.SendBack <- messageBack // 向client.Send发送
					}
					// 因为send_id肯定在线，所以这里在后端进行在线回显message，其实优化的话前端可以直接回显
					// 问题在于前后端的req和rsp结构不同，前端存储message的messageList不能存req，只能存rsp
					// 所以这里后端进行回显，前端不回显
					sendClient := k.Clients[message.SendId]
					sendClient.SendBack <- messageBack
					k.mutex.Unlock()

					// redis
					var rspString string
					rspString, err = myredis.GetKeyNilIsErr("message_list_" + message.SendId + "_" + message.ReceiveId)
					if err == nil {
						var rsp []respond.GetMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Error("failed to unmarshal message", zap.Error(err))
						}
						rsp = append(rsp, messageRsp)
						rspByte, err := json.Marshal(rsp)
						if err != nil {
							k.log.Error("failed to marshal message", zap.Error(err))
						}
						if err := myredis.SetKeyEx("message_list_"+message.SendId+"_"+message.ReceiveId, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
							k.log.Error("failed to set message", zap.Error(err))
						}
					} else {
						if !errors.Is(err, redis.Nil) {
							k.log.Error("failed to get message", zap.Error(err))
						}
					}
				} else {
					messageRsp := respond.GetGroupMessageListRespond{
						SendId:     message.SendId,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       message.Type,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Error("failed to marshal message", zap.Error(err))
					}
					k.log.Info("返回的消息为：", messageRsp, "序列化后为：", jsonMessage)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					var group model.GroupInfo
					if res := dao.GormDB.Where("uuid = ?", message.ReceiveId).First(&group); res.Error != nil {
						k.log.Error("failed to get group", zap.Error(res.Error))
					}
					var members []string
					if err := json.Unmarshal(group.Members, &members); err != nil {
						k.log.Error("failed to unmarshal members", zap.Error(err))
					}
					k.mutex.Lock()
					for _, member := range members {
						if member != message.SendId {
							if receiveClient, ok := k.Clients[member]; ok {
								receiveClient.SendBack <- messageBack
							}
						} else {
							sendClient := k.Clients[message.SendId]
							sendClient.SendBack <- messageBack
						}
					}
					k.mutex.Unlock()

					// redis
					var rspString string
					rspString, err = myredis.GetKeyNilIsErr("group_messagelist_" + message.ReceiveId)
					if err == nil {
						var rsp []respond.GetGroupMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Error("failed to unmarshal message", zap.Error(err))
						}
						rsp = append(rsp, messageRsp)
						rspByte, err := json.Marshal(rsp)
						if err != nil {
							k.log.Error("failed to marshal message", zap.Error(err))
						}
						if err := myredis.SetKeyEx("group_messagelist_"+message.ReceiveId, string(rspByte), time.Minute*constants.REDIS_TIMEOUT); err != nil {
							k.log.Error("failed to set message", zap.Error(err))
						}
					} else {
						if !errors.Is(err, redis.Nil) {
							k.log.Error("failed to get message", zap.Error(err))
						}
					}
				}
			} else if chatMessageReq.Type == message_type_enum.AudioOrVideo {
				var avData request.AVData
				if err := json.Unmarshal([]byte(chatMessageReq.AVdata), &avData); err != nil {
					k.log.Error("failed to unmarshal av data", zap.Error(err))
				}
				//log.Println(avData)
				message := model.Message{
					Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type,
					Content:    "",
					Url:        "",
					SendId:     chatMessageReq.SendId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: chatMessageReq.SendAvatar,
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   "",
					FileType:   "",
					FileName:   "",
					Status:     message_status_enum.Unsent,
					CreatedAt:  time.Now(),
					AVdata:     chatMessageReq.AVdata,
				}
				if avData.MessageId == "PROXY" && (avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
					// 存message
					// 对SendAvatar去除前面/static之前的所有内容，防止ip前缀引入
					message.SendAvatar = normalizePath(message.SendAvatar)
					if res := dao.GormDB.Create(&message); res.Error != nil {
						k.log.Error("failed to create message", zap.Error(res.Error))
					}
				}

				if chatMessageReq.ReceiveId[0] == 'U' { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := respond.AVMessageRespond{
						SendId:     message.SendId,
						SendName:   message.SendName,
						SendAvatar: message.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       message.Type,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
						AVdata:     message.AVdata,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Error("failed to marshal message", zap.Error(err))
					}
					// log.Println("返回的消息为：", messageRsp, "序列化后为：", jsonMessage)
					k.log.Info("返回的消息为：", messageRsp)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					k.mutex.Lock()
					if receiveClient, ok := k.Clients[message.ReceiveId]; ok {
						//messageBack.Message = jsonMessage
						//messageBack.Uuid = message.Uuid
						receiveClient.SendBack <- messageBack // 向client.Send发送
					}
					// 通话这不能回显，发回去的话就会出现两个start_call。
					//sendClient := s.Clients[message.SendId]
					//sendClient.SendBack <- messageBack
					k.mutex.Unlock()
				}
			}
		}
	}()

	// login, logout message
	for {
		select {
		case client := <-k.Login:
			{
				k.mutex.Lock()
				k.Clients[client.Uuid] = client
				k.mutex.Unlock()
				k.log.Info(fmt.Sprintf("欢迎来到kama聊天服务器，亲爱的用户%s\n", client.Uuid))
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到kama聊天服务器"))
				if err != nil {
					k.log.Error("failed to write message", zap.Error(err))
				}
			}

		case client := <-k.Logout:
			{
				k.mutex.Lock()
				delete(k.Clients, client.Uuid)
				k.mutex.Unlock()
				k.log.Info(fmt.Sprintf("用户%s退出登录\n", client.Uuid))
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					k.log.Error("failed to write message", zap.Error(err))
				}
			}
		}
	}
}

func (k *KafkaServerUseCase) Close() {
	close(k.Login)
	close(k.Logout)
}

func (k *KafkaServerUseCase) SendClientToLogin(client *Client) {
	k.mutex.Lock()
	k.Login <- client
	k.mutex.Unlock()
}

func (k *KafkaServerUseCase) SendClientToLogout(client *Client) {
	k.mutex.Lock()
	k.Logout <- client
	k.mutex.Unlock()
}

func (k *KafkaServerUseCase) RemoveClient(uuid string) {
	k.mutex.Lock()
	delete(k.Clients, uuid)
	k.mutex.Unlock()
}
