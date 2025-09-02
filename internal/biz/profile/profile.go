package profile

import (
	"context"
	"time"
)

type ProfileTB struct {
	ID     uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserID uint32 `gorm:"column:user_id;type:int(10) unsigned;not null;uniqueIndex;comment:关联的用户ID" json:"user_id"`

	Tags string `gorm:"column:tags;type:varchar(255);comment:用户标签/兴趣;default:null" json:"tags"`

	FollowCount uint32 `gorm:"column:follow_count;type:int(10) unsigned;default:0;comment:关注数量" json:"follow_count"`
	FanCount    uint32 `gorm:"column:fan_count;type:int(10) unsigned;default:0;comment:粉丝数量" json:"fan_count"`

	ViewCount         uint32 `gorm:"column:view_count;type:int(10) unsigned;default:0;comment:主页浏览次数" json:"view_count"`
	NoteCount         uint32 `gorm:"column:note_count;type:int(10) unsigned;default:0;comment:笔记数量" json:"note_count"`
	ReceivedLikeCount uint32 `gorm:"column:received_like_count;type:int(10) unsigned;default:0;comment:获得点赞数" json:"received_like_count"`
	CollectedCount    uint32 `gorm:"column:collected_count;type:int(10) unsigned;default:0;comment:获得收藏数" json:"collected_count"`
	CommentCount      uint32 `gorm:"column:comment_count;type:int(10) unsigned;default:0;comment:评论数量" json:"comment_count"`

	// 扩展字段
	LastLoginIP string     `gorm:"column:last_login_ip;type:varchar(45);comment:最后登录IP;default:null" json:"last_login_ip"`
	LastActive  *time.Time `gorm:"column:last_active;type:datetime;default null;comment:最后活跃时间" json:"last_active"`
	Status      string     `gorm:"column:status;type:varchar(20);default:'active';comment:用户状态 active/ban/etc" json:"status"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

type FollowTB struct {
	ID uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`

	// 关注者ID（谁去关注别人）
	FollowerID uint32 `gorm:"column:follower_id;type:int(10) unsigned;not null;uniqueIndex:idx_follower_followee;comment:关注者ID" json:"follower_id"`

	// 被关注者ID（谁被关注）
	FolloweeID uint32 `gorm:"column:followee_id;type:int(10) unsigned;not null;uniqueIndex:idx_follower_followee;comment:被关注者ID" json:"followee_id"`

	// 状态，normal=正常，deleted=取消关注、拉黑等
	Status string `gorm:"column:status;type:varchar(20);default:'normal';comment:关注状态 normal/cancel" json:"status"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

func (p *ProfileTB) TableName() string {
	return "t_user_profile"
}

func (f *FollowTB) TableName() string {
	return "t_user_follow_relationships"
}

type ProfileRepo interface {
	CreateProfile(ctx context.Context, profile *ProfileTB) error
	GetProfileByUserID(ctx context.Context, userID uint32) (*ProfileTB, error)
	UpdateProfile(ctx context.Context, profile *ProfileTB) error

	FollowUser(ctx context.Context, followerID uint32, followeeID uint32) error
	CheckFollowTogether(ctx context.Context, followerID uint32, followeeID uint32) (bool, error)

	// 增量更新统计字段
	IncrementFollowCount(ctx context.Context, userID uint, delta int) error
	IncrementFanCount(ctx context.Context, userID uint, delta int) error

	//IncrementViewCount(ctx context.Context, userID uint, delta int) error
	//IncrementNoteCount(ctx context.Context, userID uint, delta int) error
	//IncrementReceivedLikeCount(ctx context.Context, userID uint, delta int) error
	//IncrementCollectedCount(ctx context.Context, userID uint, delta int) error
	//IncrementCommentCount(ctx context.Context, userID uint, delta int) error
}
