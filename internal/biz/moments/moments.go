package moments

import (
	"context"
	"time"
)

type MomentsBoxTB struct {
	ID         uint32          `gorm:"primarykey"`
	UserID     uint32          `gorm:"column:user_id;comment:用户ID"`
	SendBox    []MomentsMetaTB `gorm:"foreignKey:UserID;references:UserID;comment:发送盒"`
	ReceiveBox []MomentsMetaTB `gorm:"foreignKey:UserID;references:UserID;comment:接收盒"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (m *MomentsBoxTB) TableName() string {
	return "t_moments_box"
}

type MomentsBoxRepo interface {
	CreateMomentsBox(ctx context.Context, momentsBox *MomentsBoxTB) error
	DeleteMomentsBox(ctx context.Context, momentsBoxID uint32) error
}
