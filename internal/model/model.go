package model

import (
	"github.com/google/wire"
	"kratos-realworld/internal/model/infra"
)

// ProviderSet is model providers.
var ProviderSet = wire.NewSet(
	// 最底层 model 里面定义通用的Data结构体以及需要操作data的一些接口
	NewData,
	NewTransaction,

	// infra里面初始化数据库和redis
	infra.NewDatabase,
	infra.NewCache,
)
