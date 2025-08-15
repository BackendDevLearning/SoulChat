package service

import (
	"github.com/google/wire"
	bizUser "kratos-realworld/internal/biz/user"

	"github.com/go-kratos/kratos/v2/log"
	v1 "kratos-realworld/api/conduit/v1"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewConduitService)

type ConduitService struct {
	v1.UnimplementedConduitServer

	ur  *bizUser.UserRegisterCase
	ul  *bizUser.UserLoginCase
	log *log.Helper
}

func NewConduitService(ul *bizUser.UserLoginCase, ur *bizUser.UserRegisterCase, logger log.Logger) *ConduitService {
	return &ConduitService{
		ul:  ul,
		ur:  ur,
		log: log.NewHelper(logger)}
}

func (s *ConduitService) UR() *bizUser.UserRegisterCase {
	return s.ur
}
