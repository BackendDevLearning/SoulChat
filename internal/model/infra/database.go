package infra

import (
	"fmt"
	"gorm.io/gorm"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/model/gormcli"
)

func NewDatabase(conf *conf.Data) *gorm.DB {
	fmt.Println("NewDatabase")
	dt := conf.GetDatabase()
	gormcli.Init(
		gormcli.WithAddr(dt.GetAddr()),
		gormcli.WithUser(dt.GetUser()),
		gormcli.WithPassword(dt.GetPassword()),
		gormcli.WithDataBase(dt.GetDatabase()),
		gormcli.WithMaxIdleConn(int(dt.GetMaxIdleConn())),
		gormcli.WithMaxOpenConn(int(dt.GetMaxOpenConn())),
		gormcli.WithMaxIdleTime(int64(dt.GetMaxIdleTime())),
		// 如果设置了慢查询阈值，就打印日志
		gormcli.WithSlowThresholdMillisecond(dt.GetSlowThresholdMillisecond()),
	)

	return gormcli.GetDB()
}
