package sql

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/data"
)

var ProviderSet = wire.NewSet(NewData, NewDB, data.NewUserRepo, data.NewProfileRepo, data.NewArticleRepo, data.NewCommentRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{db: db}, cleanup, nil
}

func NewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}
	InitDB(db)
	return db
}

func InitDB(db *gorm.DB) {
	if err := db.AutoMigrate(
		&data.User{},
		&data.Article{},
		&data.Comment{},
		&data.ArticleFavorite{},
		&data.Following{},
	); err != nil {
		panic(err)
	}
}
