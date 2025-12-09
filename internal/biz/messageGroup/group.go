package messageGroup

import (
	"time"
	"kratos-realworld/internal/common/res"
)

type GroupTB struct {
	ID     uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid   string `gorm:"column:uuid;type:varchar(150);not null;uniqueIndex:idx_uuid;comment:uuid" json:"uuid"`
	CreaterID uint32 `gorm:"column:user_id;type:int(10) unsigned;not null;index;comment:群主ID" json:"userId"`
	Member string `gorm:"column:member;type:text;comment:群成员列表(JSON数组)" json:"member"`
	Adminer string `gorm:"column:adminer;type:text;comment:管理员ID(JSON数组)" json:"adminer"`
	Name   string `gorm:"column:name;type:varchar(150);not null;comment:群名称" json:"name"`
	Notice string `gorm:"column:notice;type:varchar(350);comment:群公告" json:"notice"`
	Mode   uint32 `gorm:"column:mode;type:int(10) unsigned;not null;default:0;comment:群模式 0-公开群 1-私密群" json:"mode"`
	AddMode uint32 `gorm:"column:add_mode;type:int(10) unsigned;not null;default:0;comment:加群方式 0-自由加入 1-需要验证" json:"add_mode"`
	Avatar string `gorm:"column:avatar;type:varchar(250);comment:群头像" json:"avatar"`
	Intro  string `gorm:"column:intro;type:varchar(500);comment:群简介" json:"intro"`
	MemberCount uint32 `gorm:"column:member_count;type:int(10) unsigned;not null;default:0;comment:群成员数量" json:"member_count"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (g *GroupTB) TableName() string {
	return "t_group"
}

type GroupInfoRepo interface {
	CreateGroup(user_id uint32, name string, mode uint32, add_mode uint32, intro string) (uint32, error)
	LoadMyGroup(UserId uint32) ([]res.LoadMyGroupData, error)
	LoadJoinGroup(UserId uint32) ([]res.LoadMyGroupData, error)
	SetAdmin(UserId uint32, GroupId uint32) error
}
