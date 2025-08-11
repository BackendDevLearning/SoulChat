package service

import (
	"github.com/google/wire"
	datauser "kratos-realworld/internal/data/user"

	"github.com/go-kratos/kratos/v2/log"
	v1 "kratos-realworld/api/conduit/v1"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewConduitService)

type ConduitService struct {
	v1.UnimplementedConduitServer

	uc  *datauser.UserLogRepo
	log *log.Helper
}

func NewConduitService(uc *datauser.UserLogRepo, logger log.Logger) *ConduitService {
	return &ConduitService{uc: uc, log: log.NewHelper(logger)}
}
