package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/common/req"
	"kratos-realworld/internal/common/res"
	"kratos-realworld/internal/kafka"
	"kratos-realworld/internal/model"
	"os"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)


// AVData 音视频数据
type AVData struct {
	MessageId string `json:"messageId"`
	Type      string `json:"type"`
}

// formatMessageTime 格式化消息时间，如果为 nil 则使用当前时间
func formatMessageTime(t *time.Time, defaultTime time.Time) string {
	if t != nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return defaultTime.Format("2006-01-02 15:04:05")
}

type KafkaServerUseCase struct {
	Clients      map[string]*Client
	mutex        *sync.Mutex
	Login        chan *Client // 登录通道
	Logout       chan *Client // 退出登录通道
	log          *log.Helper
	data         *model.Data
	kafkaService *kafka.KafkaService
}

// 用来接收操作系统的信号
var kafkaQuit = make(chan os.Signal, 1)

func NewKafkaServerUseCase(log *log.Helper, data *model.Data, kafkaService *kafka.KafkaService) *KafkaServerUseCase {
	return &KafkaServerUseCase{
		Clients:      make(map[string]*Client),
		mutex:        &sync.Mutex{},
		Login:        make(chan *Client),
		Logout:       make(chan *Client),
		log:          log,
		data:         data,
		kafkaService: kafkaService,
	}
}

// Start 启动 Kafka 服务器
func (k *KafkaServerUseCase) Start() {
	defer func() {
		if r := recover(); r != nil {
			if k.log != nil {
				k.log.Errorf("kafka server panic, %v", r)
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
					k.log.Errorf("kafka server panic in goroutine, %v", r)
				}
			}
		}()
		ctx := context.Background()
		for {
			if kafka.KafkaService.ChatReader == nil {
				if k.log != nil {
					k.log.Warn("kafka reader is not initialized")
				}
				time.Sleep(time.Second)
				continue
			}

			kafkaMessage, err := k.kafkaService.ChatReader.ReadMessage(ctx)
			if err != nil {
				if k.log != nil {
					k.log.Errorf("failed to read kafka message, %v", err)
				}
				continue
			}

			if k.log != nil {
				k.log.Infof("received kafka message, topic=%s, partition=%d, offset=%d, key=%s, value=%s",
					kafkaMessage.Topic,
					kafkaMessage.Partition,
					kafkaMessage.Offset,
					string(kafkaMessage.Key),
					string(kafkaMessage.Value),
				)
			}

			data := kafkaMessage.Value
			var chatMessageReq req.ChatMessageRequest
			if err := json.Unmarshal(data, &chatMessageReq); err != nil {
				if k.log != nil {
					k.log.Errorf("failed to unmarshal chat message, %v", err)
				}
				continue
			}

			k.log.Infof("原消息为：%s, 反序列化后为：%+v", string(data), chatMessageReq)

			if chatMessageReq.Type == MessageTypeText {
				// 存message
				now := time.Now()
				message := messageGroup.MessageTB{
					Uuid:       GenerateMessageUUID(),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type,
					Content:    chatMessageReq.Content,
					Url:        "",
					FromUserID: chatMessageReq.SendId,
					ToUserID:   chatMessageReq.ReceiveId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: normalizePath(chatMessageReq.SendAvatar),
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   "0B",
					FileType:   "",
					FileName:   "",
					Status:     common.MessageStatusUnsent, // 0.未发送
					MessageType: chatMessageReq.MessageType, // 1单聊
					AVdata:     "",
					CreatedAt:  &now,
				}
				// 判断是单聊还是群聊
				if err := k.data.DB().WithContext(ctx).Create(&message).Error; err != nil {
					k.log.Errorf("failed to create message, %v", err)
				}
				
				if chatMessageReq.MessageType == common.MessageTypeUser { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := res.GetMessageListRespond{
						SendId:     message.FromUserID,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       common.MessageTypeText,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  formatMessageTime(message.CreatedAt, now),
						MessageType: message.MessageType,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Errorf("failed to marshal message, %v", err)
					}
					k.log.Infof("返回的消息为：%+v, 序列化后为：%s", messageRsp, string(jsonMessage))
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
					key := "message_list_" + message.FromUserID + "_" + message.ReceiveId
					rspString, exists, err := k.data.Cache().Get(ctx, key)
					if err == nil && exists {
						var rsp []res.GetMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Errorf("failed to unmarshal message, %v", err)
						} else {
							rsp = append(rsp, messageRsp)
							rspByte, err := json.Marshal(rsp)
							if err != nil {
								k.log.Errorf("failed to marshal message, %v", err)
							} else {
								if err := k.data.Cache().Set(ctx, key, string(rspByte), time.Minute*common.REDIS_TIMEOUT); err != nil {
									k.log.Errorf("failed to set message, %v", err)
								}
							}
						}
					} else if err != nil && !errors.Is(err, redis.Nil) {
						k.log.Errorf("failed to get message, %v", err)
					}

				} else if message.MessageType == common.MessageTypeGroup { // 发送给Group
					messageRsp := res.GetGroupMessageListRespond{
						SendId:     message.FromUserID,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       common.MessageTypeText,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  formatMessageTime(message.CreatedAt, now),
						MessageType: message.MessageType,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Errorf("failed to marshal message, %v", err)
					}
					k.log.Infof("返回的消息为：%+v, 序列化后为：%s", messageRsp, string(jsonMessage))
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					// 查询群组信息
					var group messageGroup.GroupTB
					if err := k.data.DB().WithContext(ctx).Where("uuid = ?", message.ReceiveId).First(&group).Error; err != nil {
						k.log.Errorf("failed to get group, %v", err)
						continue
					}
					// 查询群成员
					var groupMembers []messageGroup.GroupMemberTB
					if err := k.data.DB().WithContext(ctx).Where("group_id = ?", group.ID).Find(&groupMembers).Error; err != nil {
						k.log.Errorf("failed to get group members, %v", err)
						continue
					}
					var members []string
					for _, member := range groupMembers {
						members = append(members, fmt.Sprintf("U%d", member.UserID))
					}
					k.mutex.Lock()
					for _, member := range members {
						if member != message.FromUserID {
							if receiveClient, ok := k.Clients[member]; ok {
								receiveClient.SendBack <- messageBack
							}
						} else {
							if sendClient, ok := k.Clients[message.FromUserID]; ok {
								sendClient.SendBack <- messageBack
							}
						}
					}
					k.mutex.Unlock()

					// redis
					key := "group_messagelist_" + message.ReceiveId
					rspString, exists, err := k.data.Cache().Get(ctx, key)
					if err == nil && exists {
						var rsp []res.GetGroupMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Errorf("failed to unmarshal message, %v", err)
						} else {
							rsp = append(rsp, messageRsp)
							rspByte, err := json.Marshal(rsp)
							if err != nil {
								k.log.Errorf("failed to marshal message, %v", err)
							} else {
								if err := k.data.Cache().Set(ctx, key, string(rspByte), time.Minute*common.REDIS_TIMEOUT); err != nil {
									k.log.Errorf("failed to set message, %v", err)
								}
							}
						}
					} else if err != nil && !errors.Is(err, redis.Nil) {
						k.log.Errorf("failed to get message, %v", err)
					}
				}
			} else if chatMessageReq.Type == common.MessageTypeFile {
				// 存message
				now := time.Now()
				message := messageGroup.MessageTB{
					Uuid:       GenerateMessageUUID(),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type, // 2.文件
					Content:    "",
					Url:        chatMessageReq.Url,
					FromUserID: chatMessageReq.SendId,
					ToUserID:   chatMessageReq.ReceiveId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: normalizePath(chatMessageReq.SendAvatar),
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   chatMessageReq.FileSize,
					FileType:   chatMessageReq.FileType,
					FileName:   chatMessageReq.FileName,
					Status:     common.MessageStatusUnsent, // 0.未发送
					MessageType: chatMessageReq.MessageType, // 1单聊
					AVdata:     "",
					CreatedAt:  formatMessageTime(message.CreatedAt, now),
				}
				
				if err := k.data.DB().WithContext(ctx).Create(&message).Error; err != nil {
					k.log.Errorf("failed to create message, %v", err)
				}
				
				if chatMessageReq.MessageType == common.MessageTypeUser { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := res.GetMessageListRespond{
						SendId:     message.FromUserID,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       common.MessageTypeFile,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  formatMessageTime(message.CreatedAt, now),
						MessageType: message.MessageType,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Errorf("failed to marshal message, %v", err)
					}
					k.log.Infof("返回的消息为：%+v, 序列化后为：%s", messageRsp, string(jsonMessage))
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					k.mutex.Lock()
					if receiveClient, ok := k.Clients[message.ReceiveId]; ok {
						receiveClient.SendBack <- messageBack
					}
					if sendClient, ok := k.Clients[message.FromUserID]; ok {
						sendClient.SendBack <- messageBack
					}
					k.mutex.Unlock()

					// redis
					key := "message_list_" + message.FromUserID + "_" + message.ReceiveId
					rspString, exists, err := k.data.Cache().Get(ctx, key)
					if err == nil && exists {
						var rsp []res.GetMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Errorf("failed to unmarshal message, %v", err)
						} else {
							rsp = append(rsp, messageRsp)
							rspByte, err := json.Marshal(rsp)
							if err != nil {
								k.log.Errorf("failed to marshal message, %v", err)
							} else {
								if err := k.data.Cache().Set(ctx, key, string(rspByte), time.Minute*common.REDIS_TIMEOUT); err != nil {
									k.log.Errorf("failed to set message, %v", err)
								}
							}
						}
					} else if err != nil && !errors.Is(err, redis.Nil) {
						k.log.Errorf("failed to get message, %v", err)
					}
				} else if chatMessageReq.MessageType == common.MessageTypeGroup { // 发送给Group
					messageRsp := res.GetGroupMessageListRespond{
						SendId:     message.FromUserID,
						SendName:   message.SendName,
						SendAvatar: chatMessageReq.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       common.MessageTypeFile,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  formatMessageTime(message.CreatedAt, now),
						MessageType: chatMessageReq.MessageType,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Errorf("failed to marshal message, %v", err)
					}
					k.log.Infof("返回的消息为：%+v, 序列化后为：%s", messageRsp, string(jsonMessage))
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					// 查询群组信息
					var group messageGroup.GroupTB
					if err := k.data.DB().WithContext(ctx).Where("uuid = ?", message.ReceiveId).First(&group).Error; err != nil {
						k.log.Errorf("failed to get group, %v", err)
						continue
					}
					// 查询群成员
					var groupMembers []messageGroup.GroupMemberTB
					if err := k.data.DB().WithContext(ctx).Where("group_id = ?", group.ID).Find(&groupMembers).Error; err != nil {
						k.log.Errorf("failed to get group members, %v", err)
						continue
					}
					var members []string
					for _, member := range groupMembers {
						members = append(members, fmt.Sprintf("U%d", member.UserID))
					}
					k.mutex.Lock()
					for _, member := range members {
						if member != message.FromUserID {
							if receiveClient, ok := k.Clients[member]; ok {
								receiveClient.SendBack <- messageBack
							}
						} else {
							if sendClient, ok := k.Clients[message.FromUserID]; ok {
								sendClient.SendBack <- messageBack
							}
						}
					}
					k.mutex.Unlock()

					// redis
					key := "group_messagelist_" + message.ReceiveId
					rspString, exists, err := k.data.Cache().Get(ctx, key)
					if err == nil && exists {
						var rsp []res.GetGroupMessageListRespond
						if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
							k.log.Errorf("failed to unmarshal message, %v", err)
						} else {
							rsp = append(rsp, messageRsp)
							rspByte, err := json.Marshal(rsp)
							if err != nil {
								k.log.Errorf("failed to marshal message, %v", err)
							} else {
								if err := k.data.Cache().Set(ctx, key, string(rspByte), time.Minute*common.REDIS_TIMEOUT); err != nil {
									k.log.Errorf("failed to set message, %v", err)
								}
							}
						}
					} else if err != nil && !errors.Is(err, redis.Nil) {
						k.log.Errorf("failed to get message, %v", err)
					}
				}
			// smc todo:通话这里需要好好看看怎么实现的
			} else if chatMessageReq.Type == common.MessageTypeAudioOrVideo {
				var avData AVData
				if err := json.Unmarshal([]byte(chatMessageReq.AVdata), &avData); err != nil {
					k.log.Errorf("failed to unmarshal av data, %v", err)
				}
				now := time.Now()
				message := messageGroup.MessageTB{
					Uuid:       GenerateMessageUUID(),
					SessionId:  chatMessageReq.SessionId,
					Type:       chatMessageReq.Type, // 3.通话
					Content:    "",
					Url:        "",
					FromUserID: chatMessageReq.SendId,
					ToUserID:   chatMessageReq.ReceiveId,
					SendName:   chatMessageReq.SendName,
					SendAvatar: normalizePath(chatMessageReq.SendAvatar),
					ReceiveId:  chatMessageReq.ReceiveId,
					FileSize:   "",
					FileType:   "",
					FileName:   "",
					Status:     common.MessageStatusUnsent, // 0.未发送
					MessageType: chatMessageReq.MessageType, // 1单聊
					AVdata:     chatMessageReq.AVdata,
					CreatedAt:  formatMessageTime(message.CreatedAt, now),
				}
				
				if avData.MessageId == "PROXY" && (avData.Type == "start_call" || avData.Type == "receive_call" || avData.Type == "reject_call") {
					// 存message
					if err := k.data.DB().WithContext(ctx).Create(&message).Error; err != nil {
						k.log.Errorf("failed to create message, %v", err)
					}
				}

				if chatMessageReq.MessageType == common.MessageTypeUser { // 发送给User
					// 如果能找到ReceiveId，说明在线，可以发送，否则存表后跳过
					// 因为在线的时候是通过websocket更新消息记录的，离线后通过存表，登录时只调用一次数据库操作
					// 切换chat对象后，前端的messageList也会改变，获取messageList从第二次就是从redis中获取
					messageRsp := res.AVMessageRespond{
						SendId:     message.FromUserID,
						SendName:   message.SendName,
						SendAvatar: message.SendAvatar,
						ReceiveId:  message.ReceiveId,
						Type:       common.MessageTypeAudioOrVideo,
						Content:    message.Content,
						Url:        message.Url,
						FileSize:   message.FileSize,
						FileName:   message.FileName,
						FileType:   message.FileType,
						CreatedAt:  formatMessageTime(message.CreatedAt, now),
						AVdata:     message.AVdata,
						MessageType: message.MessageType,
					}
					jsonMessage, err := json.Marshal(messageRsp)
					if err != nil {
						k.log.Errorf("failed to marshal message, %v", err)
					}
					k.log.Infof("返回的消息为：%+v", messageRsp)
					var messageBack = &MessageBack{
						Message: jsonMessage,
						Uuid:    message.Uuid,
					}
					k.mutex.Lock()
					if receiveClient, ok := k.Clients[message.ReceiveId]; ok {
						receiveClient.SendBack <- messageBack
					}
					// 通话这不能回显，发回去的话就会出现两个start_call。
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
				k.log.Infof("欢迎来到kama聊天服务器，亲爱的用户%s", client.Uuid)
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte("欢迎来到kama聊天服务器"))
				if err != nil {
					k.log.Errorf("failed to write message, %v", err)
				}
			}

		case client := <-k.Logout:
			{
				k.mutex.Lock()
				delete(k.Clients, client.Uuid)
				k.mutex.Unlock()
				k.log.Infof("用户%s退出登录", client.Uuid)
				if err := client.Conn.WriteMessage(websocket.TextMessage, []byte("已退出登录")); err != nil {
					k.log.Errorf("failed to write message, %v", err)
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
