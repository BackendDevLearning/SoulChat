package data

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/model"
)

type MessageRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewMessageRepo(data *model.Data, logger log.Logger) messageGroup.MessageRepo {
	return &MessageRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (mr *MessageRepo) GetMessages(message common.MessageRequest) ([]common.MessageResponse, error) {
	//	todo:
	fmt.Print("todo")
	return nil, nil
}

func (mr *MessageRepo) FetchGroupMessage(toUuid string) ([]common.MessageResponse, error) {
	return nil, nil
}

func (mr *MessageRepo) SaveMessage(message v1.Message) error {
	// TODO: implement message saving logic
	return nil
}
