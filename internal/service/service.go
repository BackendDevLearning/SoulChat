package service

import (
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewConduitService)

type ConduitService struct {
	v1.UnimplementedConduitServer

	gt  *biz.GateWayUsecase
	pc  *biz.ProfileUsecase
	log *log.Helper
}

func NewConduitService(gt *biz.GateWayUsecase, pc *biz.ProfileUsecase, logger log.Logger) *ConduitService {
	return &ConduitService{
		gt:  gt,
		pc:  pc,
		log: log.NewHelper(logger)}
}
