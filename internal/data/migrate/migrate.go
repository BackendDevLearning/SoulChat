package migrate

import (
	"gorm.io/gorm"
	"kratos-realworld/internal/data"
)

func InitDBTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&data.User{},
		&data.Article{},
		&data.Comment{},
		&data.ArticleFavorite{},
		&data.Following{},
	); err != nil {
		return err
	}
	return nil
}
