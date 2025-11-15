package moment

import (
	"context"
	"time"
)

type MomentTB struct {
	ID            uint32       `gorm:"primarykey"`
	UserID        uint32       `gorm:"column:user_id;comment:发送动态的用户ID"`
	Message       string       `gorm:"type:varchar(500);column:message;comment:动态内容"`
	MediaURL      string       `gorm:"type:varchar(500);column:media_url;comment:动态媒体URL"`
	ReceiveBoxIDs []uint32     `gorm:"column:receive_box_ids;comment:接收盒IDs"`
	Comments      []CommentsTB `gorm:"foreignKey:MomentID;references:ID"`
	LikeCount     int          `gorm:"column:like_count;comment:点赞数"`
	LikeIDs       []uint32     `gorm:"column:like_ids;comment:点赞用户IDs"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (m *MomentTB) TableName() string {
	return "t_moments"
}

type MomentRepo interface {
	CreateMoment(ctx context.Context, moment *MomentTB) error
	DeleteMoment(ctx context.Context, momentID uint32) (*MomentTB, error)
	GetMoment(ctx context.Context, momentID uint32) (*MomentTB, error)
	CreateComment(ctx context.Context, comment *CommentsTB) error
	DeleteComment(ctx context.Context, commentID uint32) error
	GetMomentMeta(ctx context.Context, momentID uint32) (*MomentsMetaTB, error)
	GetComments(ctx context.Context, momentID uint32) ([]*CommentsTB, error)
}
