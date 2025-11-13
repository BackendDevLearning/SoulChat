package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
)

type MessageUseCase struct {
	mr  bizChat.MessageRepo
	log *log.Helper
}

func NewMessageUseCase(mr bizChat.MessageRepo, logger log.Logger) *MessageUseCase {
	return &MessageUseCase{
		mr:  mr,
		log: log.NewHelper(logger),
	}
}

func (mc *MessageUseCase) SaveMessage(message *bizChat.MessageTB) error {
	err := mc.mr.SaveMessage(message)
	if err != nil {
		return NewErr(ErrCodeMessageFailed, MESSAGE_FAILED, "Save message to database failed")
	}
	return nil
}

func (mc *MessageUseCase) GetMessages(ctx context.Context, messageReq common.MessageRequest) error {
	MessageResponse, err := mc.mr.GetMessages(ctx, messageReq)
	fmt.Println(MessageResponse)
	
	if err != nil {
		return NewErr(ErrCodeMessageFailed, MESSAGE_FAILED, "Get message failed")
	}

	return nil
}

func (mc *MessageUseCase) fetchGroupMessage() {

}
