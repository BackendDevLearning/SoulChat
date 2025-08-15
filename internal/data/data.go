package data

import (
	"github.com/google/wire"
	"kratos-realworld/internal/data/user"
)

var ProviderSet = wire.NewSet(
	user.NewUserRegisterRepo,
	user.NewUserLoginRepo,
)
