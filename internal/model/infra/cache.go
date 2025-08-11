package infra

import (
	"fmt"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/model/cache"
)

func NewCache(conf *conf.Data) *cache.Client {
	fmt.Println("NewCache")
	dt := conf.GetRedis()
	cache.Init(
		cache.WithAddr(dt.GetAddr()),
		cache.WithPassWord(dt.GetPassword()),
		cache.WithDB(int(dt.GetDb())),
		cache.WithPoolSize(int(dt.GetPoolSize())),
	)

	return cache.GetRedisCli()
}
