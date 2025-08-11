package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"kratos-realworld/internal/data/user"
	"kratos-realworld/internal/model"
)

var ProviderSet = wire.NewSet(
	user.NewUserLogRepo,
)

// NewUserLogRepoProvider creates a UserLogRepo with dependencies
func NewUserLogRepoProvider(data *model.Data, logger log.Logger) *user.UserLogRepo {
	return user.NewUserLogRepo(data, logger)
}
