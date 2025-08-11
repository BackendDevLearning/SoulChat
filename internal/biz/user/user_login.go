package user

import (
	"context"
	"time"
)

// BlackIp ip黑明单表
type UserLog struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserName   string     `gorm:"column:UserName;type:varchar(50);comment:账号;NOT NULL" json:"UserName"`
	PassWord   string     `gorm:"column:PassWord;type:varchar(50);comment:密码;NOT NULL" json:"PassWord"`
	Token      string     `gorm:"column:Token;type:varchar(50);comment:Token" json:"Token"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

func (m *UserLog) TableName() string {
	return "t_user_login"
}

type UserLogRepo interface {
	CreateUser(ctx context.Context, login *UserLog) error
	UpdateByCache(user *UserLog) error
	Load(user *UserLog) error
}
