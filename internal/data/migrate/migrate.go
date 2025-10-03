package migrate

import (
	"gorm.io/gorm"
	"kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/biz/user"
)

func InitDBTable(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&user.UserTB{},
		&profile.ProfileTB{},
		&profile.FollowFanTB{},
		&messageGroup.MessageTB{},
		&messageGroup.GroupTB{},
		&messageGroup.GroupMemberTB{},
	); err != nil {
		return err
	}
	return nil
}
