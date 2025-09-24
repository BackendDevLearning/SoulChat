package chat

import "time"

type Message struct {
	ID          uint64     `gorm:"column:id;type:bigint unsigned;primary_key;AUTO_INCREMENT" json:"id"`                 // 消息ID
	CreatedAt   *time.Time `gorm:"column:created_at;type:datetime(3);default:null;comment:创建时间" json:"created_at"`      // 创建时间
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:datetime(3);default:null;comment:更新时间" json:"updated_at"`      // 更新时间
	DeletedAt   *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
	FromUserID  uint32     `gorm:"column:from_user_id;type:int(10) unsigned;comment:发送人ID" json:"from_user_id"`         // 发送人ID
	ToUserID    uint32     `gorm:"column:to_user_id;type:int(10) unsigned;comment:发送对象ID" json:"to_user_id"`            // 发送对象ID
	Content     string     `gorm:"column:content;type:varchar(2500);default:null;comment:消息内容" json:"content"`          // 消息内容
	Url         string     `gorm:"column:url;type:varchar(350);default:null;comment:文件或图片地址" json:"url"`                // 文件或图片地址
	Pic         string     `gorm:"column:pic;type:text;default:null;comment:缩略图" json:"pic"`                            // 缩略图
	MessageType int16      `gorm:"column:message_type;type:smallint;default:null;comment:消息类型" json:"message_type"`     // 消息类型：1单聊，2群聊
	ContentType int16      `gorm:"column:content_type;type:smallint;default:null;comment:消息内容类型" json:"content_type"`   // 消息内容类型：1文字，2语音，3视频

	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime(3);default null;comment:创建时间" json:"sys_created"` // 创建时间
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime(3);default null;comment:更新时间" json:"sys_updated"` // 更新时间
}

func (m *Message) TableName() string {
	return "t_message"
}

type MessageRepo interface {
	GetMessages()
	fetchGroupMessage()
	SaveMessage()
}
