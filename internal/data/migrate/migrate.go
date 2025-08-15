package migrate

import (
	"gorm.io/gorm"
	"kratos-realworld/internal/biz/user"
)

func InitDBTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&user.UserRegisterTB{},
	); err != nil {
		return err
	}
	return nil
}
