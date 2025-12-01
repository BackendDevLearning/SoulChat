package messageGroup

import "time"

type GroupMemberTB struct {
	ID       uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserID   uint32 `gorm:"column:user_id;type:int(10) unsigned;not null;index;comment:用户ID" json:"userId"`
	GroupID  uint32 `gorm:"column:group_id;type:int(10) unsigned;not null;index;comment:群组ID" json:"groupId"`
	Nickname string `gorm:"column:nickname;type:varchar(350);comment:昵称" json:"nickname"`
	Mute     uint16 `gorm:"column:mute;type:smallint;not null;default:0;comment:是否禁言" json:"mute"`
	Role     uint16 `gorm:"column:role;type:smallint;not null;default:0;comment:角色 0-普通成员 1-管理员 2-群主" json:"role"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (gm *GroupMemberTB) TableName() string {
	return "t_groupMember"
}

type GroupMemberRepo interface {

}
