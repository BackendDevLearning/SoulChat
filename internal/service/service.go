package service

import (
	"github.com/google/wire"
	bizUser "kratos-realworld/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	v1 "kratos-realworld/api/conduit/v1"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewConduitService)

type ConduitService struct {
	v1.UnimplementedConduitServer

	gc  *bizUser.GateWayUsecase
	log *log.Helper
}

func NewConduitService(gc *bizUser.GateWayUsecase, logger log.Logger) *ConduitService {
	return &ConduitService{
		gc:  gc,
		log: log.NewHelper(logger)}
}
