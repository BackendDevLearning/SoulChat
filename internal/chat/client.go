package chat

import (
	"context"
	"encoding/json"
	diyLog "github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
	kafkaGo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	message "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common/req"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/kafka"
	"kratos-realworld/internal/model"
	"strconv"
)

type MessageBack struct {
	Message []byte
	Uuid    string
}

type Client struct {
	Conn     *websocket.Conn
	Uuid     string
	SendTo   chan []byte       // 给server端
	SendBack chan *MessageBack // 给前端
}

// 空的上下文
var ctx = context.Background()

// messageMode 消息模式：kafka 或 channel，默认 kafka
var messageMode = "kafka"

// SetMessageMode 设置消息模式
func SetMessageMode(mode string) {
	messageMode = mode
}

// GetMessageMode 获取消息模式
func GetMessageMode() string {
	return messageMode
}

// Read 读取websocket消息并发送给send通道
func (c *Client) Read(logger *diyLog.Helper) {
	if logger != nil {
		logger.Info("ws read goroutine start")
	}
	for {
		// 阻塞有一定隐患，因为下面要处理缓冲的逻辑，但是可以先不做优化，问题不大
		_, jsonMessage, err := c.Conn.ReadMessage() // 阻塞状态
		if err != nil {
			if logger != nil {
				logger.Error("failed to read websocket message", diyLog.Any("error", err))
			}
			return // 直接断开websocket
		} else {
			var message = req.ChatMessageRequest{}
			if err := json.Unmarshal(jsonMessage, &message); err != nil {
				if logger != nil {
					logger.Error("failed to unmarshal message", diyLog.Any("error", err))
				}
			}
			logger.Info("接受到消息为: ", jsonMessage)
			if messageMode == "channel" {
				// 如果server的转发channel没满，先把sendto中的给transmit
				// 注意：ChatServer 需要在使用前初始化
				// TODO: 需要实现 ChatServer 或使用现有的 websocket server
				//if ChatServer != nil {
				//	for len(ChatServer.Transmit) < CHANNEL_SIZE && len(c.SendTo) > 0 {
				//		sendToMessage := <-c.SendTo
				//		ChatServer.SendMessageToTransmit(sendToMessage)
				//	}
				//	// 如果server没满，sendto空了，直接给server的transmit
				//	if len(ChatServer.Transmit) < CHANNEL_SIZE {
				//		ChatServer.SendMessageToTransmit(jsonMessage)
				//		c.SendTo <- jsonMessage
				//	} else {
				//		// 否则考虑加宽channel size，或者使用kafka
				//		if err := c.Conn.WriteMessage(websocket.TextMessage, []byte("由于目前同一时间过多用户发送消息，消息发送失败，请稍后重试")); err != nil {
				//			if logger != nil {
				//				logger.Log(log.ErrorLevel, "failed to write error message", log.Error(err))
				//			}
				//		}
				//	}
				//}
			} else {
				// 这一步的目的是 "解耦消息的接收端和发送端"。
				// 简单来说，就是让"收消息"和"发消息"不要在同一个进程同步完成，避免卡死、扩展受限。
				// 服务器收，然后转发给client（由kafka处理）
				if kafka.KafkaService.ChatWriter != nil {
					// 使用默认分区 0，目前就配置了一个分区，多个怎么处理？
					partition := 0
					if err := kafka.KafkaService.ChatWriter.WriteMessages(ctx, kafkaGo.Message{
						// 整数转为字符串，再转为切片
						Key:   []byte(strconv.Itoa(partition)),
						Value: jsonMessage,
					}); err != nil {
						if logger != nil {
							logger.Log(log.ErrorLevel, "failed to write message to kafka", log.Error(err))
						}
					} else {
						if logger != nil {
							logger.Info("message sent to kafka", diyLog.String("message", string(jsonMessage)))
						}
					}
				} else {
					if logger != nil {
						logger.Log(log.WarnLevel, "kafka writer is not initialized")
					}
				}
			}
		}
	}
}

// Write 从send通道读取消息发送给websocket
func (c *Client) Write(logger *diyLog.Helper, data *model.Data) {
	if logger != nil {
		logger.Info("ws write goroutine start")
	}
	for messageBack := range c.SendBack { // 阻塞状态
		// 通过 WebSocket 发送消息
		err := c.Conn.WriteMessage(websocket.TextMessage, messageBack.Message)
		if err != nil {
			if logger != nil {
				logger.Error("failed to write websocket message", diyLog.Any("error", err))
			}
			return // 直接断开websocket
		}
		// log.Println("已发送消息：", messageBack.Message)
		// 说明顺利发送，修改状态为已发送
		// TODO: 如果需要更新消息状态，需要实现相应的数据库操作
		// 当前项目中消息模型可能不同，需要根据实际情况调整
		if data != nil && messageBack.Uuid != "" {
			// 示例：更新消息状态（需要根据实际的消息模型调整）
			if res := data.DB().Model(&message.MessageTB{}).Where("uuid = ?", messageBack.Uuid).Update("status", "sent"); res.Error != nil {
				if logger != nil {
					logger.Error("failed to update message status", zap.Error(res.Error))
				}
			}
		}
	}
}

// NewClientInit 当 server 层已经升级连接后，由用例或 handler 调用以初始化客户端
func NewClientInit(conn *websocket.Conn, userID string, kafkaConfig *conf.Data_Kafka, logger *diyLog.Helper, data *model.Data) {
	clientId := userID

	client := &Client{
		Conn:     conn,
		Uuid:     clientId,
		SendTo:   make(chan []byte, CHANNEL_SIZE),
		SendBack: make(chan *MessageBack, CHANNEL_SIZE),
	}

	// 根据配置决定使用哪种模式
	mode := GetMessageMode()
	if kafkaConfig != nil && kafkaConfig.Enabled {
		mode = "kafka"
	}

	if mode == "channel" {
		// TODO: 需要实现 ChatServer 或使用现有的 websocket server
		if ChatServer != nil {
			ChatServer.SendClientToLogin(client)
		}
	} else {
		if KafkaChatServer != nil {
			KafkaChatServer.SendClientToLogin(client)
		}
	}
	go client.Read(logger)
	go client.Write(logger, data)
	if logger != nil {
		logger.Info("ws connection established", diyLog.String("clientId", clientId))
	}
}

// ClientLogout 当接受到前端有登出消息时，会调用该函数
func ClientLogout(clientId string, kafkaConfig *conf.Data_Kafka, logger *diyLog.Helper) (string, int) {
	mode := GetMessageMode()
	if kafkaConfig != nil && kafkaConfig.Enabled {
		mode = "kafka"
	}

	var client *Client
	if ChatServer != nil {
		client = ChatServer.Clients[clientId]
	}
	if client == nil && KafkaChatServer != nil {
		client = KafkaChatServer.Clients[clientId]
	}

	if client != nil {
		if mode == "channel" {
			if ChatServer != nil {
				ChatServer.SendClientToLogout(client)
			}
		} else {
			if KafkaChatServer != nil {
				KafkaChatServer.SendClientToLogout(client)
			}
		}
		if err := client.Conn.Close(); err != nil {
			if logger != nil {
				logger.Error("failed to close websocket connection", diyLog.Any("error", err))
			}
			return "系统错误", -1
		}
		close(client.SendTo)
		close(client.SendBack)
	}
	return "退出成功", 0
}
