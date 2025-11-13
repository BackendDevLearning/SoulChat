package moments

import (
	"context"
	"time"
)

type CommentsTB struct {
	ID       uint32 `gorm:"primarykey"`
	MomentID uint32 `gorm:"column:moment_id;comment:动态ID"`
	UserID   uint32 `gorm:"column:user_id;comment:发送评论的用户ID"`
	Comment  string `gorm:"type:varchar(500);column:comment;comment:评论内容"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;not null;comment:创建时间" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;not null;comment:更新时间" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (c *CommentsTB) TableName() string {
	return "t_comments"
}

type CommentsRepo interface {
	CreateComment(ctx context.Context, comment *CommentsTB) error
	DeleteComment(ctx context.Context, commentID uint32) error
}
