package biz

import (
	"kratos-realworld/internal/biz/user"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewGatWayCase,
	user.NewUserRegisterCase,
	user.NewUserLoginCase,
)
