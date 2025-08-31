package user

import (
	"context"
	"kratos-realworld/internal/biz/profile"
	"time"
)

type UserTB struct {
	ID           uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserName     string `gorm:"column:UserName;type:varchar(50);comment:账号;NOT NULL" json:"UserName"`
	Phone        string `gorm:"column:Phone;type:varchar(20);comment:手机号码;NOT NULL" json:"Phone"`
	PasswordHash string `gorm:"column:PassWord;type:text;comment:密码;NOT NULL" json:"PassWord"`

	Bio        string `gorm:"column:Bio;type:text;comment:个人简介" json:"Bio"`
	Image      string `gorm:"column:Image;type:varchar(255);comment:头像链接" json:"Image"`
	CoverImage string `gorm:"column:CoverImage;type:varchar(255);comment:主页背景图链接" json:"CoverImage"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`

	Profile profile.ProfileTB `gorm:"foreignKey:UserID;references:ID" json:"profile"`
}

func (u *UserTB) TableName() string {
	return "t_user"
}

type UserRepo interface {
	CreateUser(ctx context.Context, userRegister *UserTB) error
	GetUserByPhone(ctx context.Context, phone string) (*UserTB, error)
	GetPasswordByPhone(ctx context.Context, phone string) (string, error)
	UpdatePassword(ctx context.Context, phone string, newPasswordHash string) error
}
