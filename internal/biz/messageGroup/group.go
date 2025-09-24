package messageGroup

import "time"

type GroupTB struct {
	ID     uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid   string `gorm:"column:uuid;type:varchar(150);not null;uniqueIndex:idx_uuid;comment:uuid" json:"uuid"`
	UserID uint32 `gorm:"column:user_id;type:int(10) unsigned;not null;index;comment:群主ID" json:"userId"`
	Name   string `gorm:"column:name;type:varchar(150);not null;comment:群名称" json:"name"`
	Notice string `gorm:"column:notice;type:varchar(350);comment:群公告" json:"notice"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (g *GroupTB) TableName() string {
	return "t_group"
}
