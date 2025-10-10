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

func (mr *MessageRepo) GetMessages(ctx context.Context, message common.MessageRequest) ([]common.MessageResponse, error) {
	//	todo:
	fmt.Print("todo")
	return nil, nil
}

func (mr *MessageRepo) FetchGroupMessage(ctx context.Context, toUuid string) ([]common.MessageResponse, error) {
	return nil, nil
}

func (mr *MessageRepo) SaveMessage(message *bizChat.MessageTB) error {
	// TODO: implement message saving logic
	return nil
}
