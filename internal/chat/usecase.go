package chat

import (
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/model"

	"github.com/google/wire"
	"github.com/gorilla/websocket"
	"github.com/go-kratos/kratos/v2/log"
)

// ProviderSet 提供给 wire 注入使用
var ProviderSet = wire.NewSet(NewChatUsecase)

type ChatUsecase struct {
	kafkaConfig *conf.Data_Kafka
	logger      *log.Helper
	data        *model.Data
	kafkaServerUseCase *KafkaServerUseCase
}

func NewChatUsecase(kafkaConfig *conf.Data_Kafka, logger *log.Helper, data *model.Data) *ChatUsecase {
	return &ChatUsecase{kafkaConfig: kafkaConfig, logger: logger, data: data, kafkaServerUseCase: NewKafkaServerUseCase(logger, data)}
}

// Login 接受已升级的 websocket 连接并交由客户端初始化处理
func (u *ChatUsecase) Login(conn *websocket.Conn, userID string) error {
	NewClientInit(conn, userID, u.kafkaConfig, u.logger, u.data)
	return nil
}

// Logout 封装原有 ClientLogout
func (u *ChatUsecase) Logout(clientId string) (string, int) {
	return ClientLogout(clientId, u.kafkaConfig, u.logger)
}
