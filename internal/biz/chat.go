package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	bizChat "kratos-realworld/internal/biz/chat"
)

type MessageRepoCase struct {
	mr  bizChat.MessageRepo
	log *log.Helper
}

func NewMessageCase(mr bizChat.MessageRepo) *MessageRepoCase {
	return &MessageRepoCase{
		mr:  mr,
		log: log.NewHelper(logger),
	}
}

func (mr *MessageRepoCase) SaveMessage() {

}

func (mr *MessageRepoCase) GetMessages() {

}

func (mr *MessageRepoCase) fetchGroupMessage() {

}
