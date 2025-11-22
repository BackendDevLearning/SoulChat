package service

import (
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/biz"
	"kratos-realworld/internal/chat"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewConduitService)

type ConduitService struct {
	v1.UnimplementedConduitServer

	gt  *biz.GateWayUsecase
	pc  *biz.ProfileUsecase
	mc  *biz.MessageUseCase
	log *log.Helper

	chatUseCase *chat.ChatUsecase
}

func NewConduitService(gt *biz.GateWayUsecase, pc *biz.ProfileUsecase, mc *biz.MessageUseCase, chatUC *chat.ChatUsecase, logger log.Logger) *ConduitService {
	return &ConduitService{
		gt:          gt,
		pc:          pc,
		mc:          mc,
		log:         log.NewHelper(logger),
		chatUseCase: chatUC,
	}
}

func (cs *ConduitService) GetMessageUseCase() *biz.MessageUseCase {
	return cs.mc
}

func (cs *ConduitService) GetChatUsecase() *chat.ChatUsecase {
	return cs.chatUseCase
}
