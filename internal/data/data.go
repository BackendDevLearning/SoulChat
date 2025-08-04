package data

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data/cache"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDatabase, NewCache, NewCouponRepo, NewPrizeRepo,
	NewResultRepo, NewBlackIpRepo, NewBlackUserRepo, NewLotteryTimesRepo, NewTransaction)

type Data struct {
	db    *gorm.DB
	cache *cache.Client
}

func NewCache(conf *conf.Data) *cache.Client {
	dt := conf.GetRedis()
	cache.Init(
		cache.WithAddr(dt.GetAddr()),
		cache.WithPassWord(dt.GetPassword()),
		cache.WithDB(int(dt.GetDb())),
		cache.WithPoolSize(int(dt.GetPoolSize())))

	return cache.GetRedisCli()
}
