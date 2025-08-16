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
	ur  *bizUser.GateWay
	log *log.Helper
}

func NewConduitService(ur *bizUser.GateWay, logger log.Logger) *ConduitService {
	return &ConduitService{
		ur:  ur,
		log: log.NewHelper(logger)}
}
