package messageGroup

import (
	"context"
	"time"

	"kratos-realworld/internal/common/req"
	"kratos-realworld/internal/common/res"
)

type MessageTB struct {
	ID        uint32 `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Uuid      string `gorm:"column:uuid;uniqueIndex;type:char(20);not null;comment:消息uuid"`
	SessionId string `gorm:"column:session_id;index;type:char(20);not null;comment:会话uuid"`

	SendId        string `gorm:"column:send_id;type:varchar(64);not null;index;comment:发送者uuid" json:"sendId"`
	SendName      string `gorm:"column:send_name;type:varchar(20);not null;comment:发送者昵称"`
	SendAvatar    string `gorm:"column:send_avatar;type:varchar(255);not null;comment:发送者头像"`
	ReceiveId     string `gorm:"column:receive_id;index;type:char(20);not null;comment:接收者uuid"`
	ReceiveAvatar string `gorm:"column:receive_avatar;type:varchar(255);not null;comment:接收者头像"`

	Type        int8   `gorm:"column:type;not null;comment:消息类型, 0.文本, 1.语音, 2.文件, 3.通话"` // 通话不用存消息内容或者url
	MessageType int8   `gorm:"column:message_type;type:smallint unsigned;not null;default:1;comment:聊天类型: 1.单聊, 2.群聊" json:"messageType"`

	Content     string `gorm:"column:content;type:varchar(2500);not null;comment:消息内容" json:"content"`
	Url         string `gorm:"column:url;type:varchar(350);comment:文件或者图片地址" json:"url"`
	Pic         string `gorm:"column:pic;type:text;comment:缩略图" json:"pic"`
	FileType string `gorm:"column:file_type;type:char(10);comment:文件类型"`
	FileName string `gorm:"column:file_name;type:varchar(50);comment:文件名"`
	FileSize string `gorm:"column:file_size;type:char(20);comment:文件大小"`
	AVdata   string `gorm:"column:av_data;type:text;comment:通话传递数据"`
	
	Status   int8   `gorm:"column:status;not null;comment:状态, 0.未发送, 1.已发送"`

	CreatedAt  *time.Time `gorm:"column:created_at;type:datetime(3);default:null;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt  *time.Time `gorm:"column:updated_at;type:datetime(3);default:null;comment:更新时间" json:"updated_at"` // 更新时间
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;comment:更新时间;NOT NULL" json:"sys_updated"`
	DeletedAt  *uint64    `gorm:"column:deleted_at;type:bigint unsigned;default:null;comment:删除时间戳" json:"deleted_at"` // 删除时间戳
}

func (m *MessageTB) TableName() string {
	return "t_message"
}

type MessageRepo interface {
	GetMessages(ctx context.Context, message req.MessageRequest) ([]res.GetMessageListRespond, int64, error) // 分页查询 1. 分页offset  2. 游标cursor
	SaveMessage(message *MessageTB) error
	GetMessagesList(ctx context.Context, uuid1 string, uuid2 string) ([]res.GetMessageListRespond, error)
}
