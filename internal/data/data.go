package data

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"gorm.io/gorm"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data/cache"
	"kratos-realworld/internal/data/gormcli"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewDatabase, NewCache, NewTransaction, // 你的 data 实例化函数
	NewUserRepo, NewProfileRepo, NewArticleRepo, NewCommentRepo)

type Data struct {
	db    *gorm.DB
	cache *cache.Client
}

type Transaction interface {
	InTx(context.Context, func(ctx context.Context) error) error
}

type contextTxKey struct{}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	//这个调用是为了把 ctx（上下文）注入到 GORM 的操作流程中
	//.Transaction(func(tx *gorm.DB) error)，这个调用是 开启一个事务块，类似于：begin, commit
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		//将 GORM 的 tx 事务对象放入 context.Context 中；
		//ontextTxKey{} 是上下文的 key（通常是一个私有结构体，避免 key 冲突）；
		//这样下游调用（比如 repo.SaveUser(ctx, user)）就可以从 ctx 中取出 tx，然后用 tx 执行数据库操作
		return fn(ctx) // 执行这个事务函数
	})
}

func (d *Data) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB) //强行断言
	if ok {
		return tx
	}
	return d.db
}

func NewData(db *gorm.DB, cache *cache.Client) *Data {
	dt := &Data{db: db, cache: cache}
	return dt
}

func NewDatabase(conf *conf.Data) *gorm.DB {
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

func NewCache(conf *conf.Data) *cache.Client {
	dt := conf.GetRedis()
	cache.Init(
		cache.WithAddr(dt.GetAddr()),
		cache.WithPassWord(dt.GetPassword()),
		cache.WithDB(int(dt.GetDb())),
		cache.WithPoolSize(int(dt.GetPoolSize())),
	)

	return cache.GetRedisCli()
}

// 也就是说，只要 *Data 实现了 InTx(ctx, fn) 方法，它就自动是一个 Transaction，返回的d本身
func NewTransaction(d *Data) Transaction {
	return d
}
