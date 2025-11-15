package moment

import (
	"time"
)

type MomentsMetaTB struct {
	ID            uint32   `gorm:"primarykey"`
	UserID        uint32   `gorm:"column:user_id"`
	MomentID      uint32   `gorm:"column:moment_id"`
	Message       string   `gorm:"type:varchar(500);column:message"`
	MediaURL      string   `gorm:"type:varchar(500);column:media_url"`
	ReceiveBoxIDs []uint32 `gorm:"column:receive_box_ids"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (m *MomentsMetaTB) TableName() string {
	return "t_moments_meta"
}
