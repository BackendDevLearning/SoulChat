//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"kratos-realworld/internal/biz"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data"
	"kratos-realworld/internal/server"
	"kratos-realworld/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// initApp init kratos application.
//func initApp(*conf.Server, *conf.Data, *conf.JWT, log.Logger) (*kratos.App, func(), error) {
//	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
//}

type CustomApp struct {
	App *kratos.App // newApp 返回 *kratos.App
	DB  *gorm.DB    // data.ProviderSet 里面提供 *gorm.DB，用于在开发环境创建database和对应的table
}

func newCustomApp(kapp *kratos.App, db *gorm.DB) *CustomApp {
	return &CustomApp{
		App: kapp,
		DB:  db,
	}
}

var CustomProviderSet = wire.NewSet(
	data.ProviderSet,    // 数据层依赖
	biz.ProviderSet,     // 业务逻辑层
	service.ProviderSet, // 接口/服务层
	server.ProviderSet,  // HTTP/GRPC 等服务启动
	newApp,              // 构造 kratos.App
	newCustomApp,        // 封装成 *CustomApp
)

func initApp(*conf.Server, *conf.Data, *conf.JWT, log.Logger) (*CustomApp, func(), error) {
	panic(wire.Build(CustomProviderSet))
}
