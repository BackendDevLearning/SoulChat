package model

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"kratos-realworld/internal/model/cache"
)

type Data struct {
	db    *gorm.DB
	cache *cache.Client
}

func NewData(db *gorm.DB, cache *cache.Client) *Data {
	fmt.Println("NewData")
	dt := &Data{db: db, cache: cache}
	return dt
}

// 给 Data 添加公开方法获取 *gorm.DB
func (d *Data) DB() *gorm.DB {
	return d.db
}

func (d *Data) Cache() *cache.Client {
	return d.cache
}

type Transaction interface {
	InTx(context.Context, func(ctx context.Context) error) error
}

type contextTxKey struct{}

var TxKey = contextTxKey{}

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

// 也就是说，只要 *Data 实现了 InTx(ctx, fn) 方法，它就自动是一个 Transaction，返回的d本身
func NewTransaction(d *Data) Transaction {
	return d
}
