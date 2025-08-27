package migrate

import (
	"gorm.io/gorm"
	"kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/biz/user"
)

func InitDBTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&user.UserTB{},
		&profile.ProfileTB{},
	); err != nil {
		return err
	}
	return nil
}
