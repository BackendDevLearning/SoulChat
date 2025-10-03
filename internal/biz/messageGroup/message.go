package messageGroup

import (
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/common"
	"time"
)

type MessageTB struct {
	ID          uint32     `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:datetime(3);default:null;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:datetime(3);default:null;comment:更新时间" json:"updated_at"` // 更新时间
	FromUserID  uint32     `gorm:"column:from_user_id;type:int(10) unsigned;not null;index;comment:发送者用户ID" json:"fromUserId"`
	ToUserID    uint32     `gorm:"column:to_user_id;type:int(10) unsigned;not null;index;comment:发送给端的ID，可为用户ID或者群ID" json:"toUserId"`
	Content     string     `gorm:"column:content;type:varchar(2500);not null;comment:消息内容" json:"content"`
	MessageType uint16     `gorm:"column:message_type;type:smallint unsigned;not null;default:1;comment:消息类型：1单聊，2群聊" json:"messageType"`
	ContentType uint16     `gorm:"column:content_type;type:smallint unsigned;not null;default:1;comment:消息内容类型：1文字 2普通文件 3图片 4音频 5视频 6语音聊天 7视频聊天" json:"contentType"`
	Pic         string     `gorm:"column:pic;type:text;comment:缩略图" json:"pic"`
	Url         string     `gorm:"column:url;type:varchar(350);comment:文件或者图片地址" json:"url"`

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;comment:更新时间;NOT NULL" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (m *MessageTB) TableName() string {
	return "t_message"
}

type MessageRepo interface {
	GetMessages(message common.MessageRequest) ([]common.MessageResponse, error) // 分页查询 1. 分页offset  2. 游标cursor
	FetchGroupMessage(toUuid string) ([]common.MessageResponse, error)
	SaveMessage(message v1.Message) error
}
