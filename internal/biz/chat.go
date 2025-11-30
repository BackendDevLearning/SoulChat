package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common/req"
	"kratos-realworld/internal/common/res"
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

func (mc *MessageUseCase) GetMessages(ctx context.Context, messageReq req.MessageRequest) ([]res.GetMessageListRespond, int64, error) {
	MessageResponse, total, err := mc.mr.GetMessages(ctx, messageReq)
	if err != nil {
		return NewErr(ErrCodeMessageFailed, MESSAGE_FAILED, "Get message failed")
	}
	return MessageResponse, total, nil
}

func (mc *MessageUseCase) fetchGroupMessage() {

}


func (mc *MessageUseCase) GetMessageList(ctx context.Context, uuid1 string, uuid2 string) ([]res.GetMessageListRespond, error) {
	res, err := mc.mr.GetMessagesList(ctx, uuid1, uuid2)
	var messageList []res.GetMessageListRespond
	if err != nil {
		mc.log.Errorf("GetMessageList err: %v\n", err)
		return []res.GetMessageListRespond{}, err
	}
	return ConvertToMessageList(res), nil
}
