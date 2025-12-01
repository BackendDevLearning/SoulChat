package req

// ChatMessageRequest 聊天消息请求结构
type ChatMessageRequest struct {
	SessionId  string `json:"session_id"`
	Type       int32   `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	SendId     string `json:"send_id"`
	SendName   string `json:"send_name"`
	SendAvatar string `json:"send_avatar"`
	ReceiveAvatar string `json:"receive_avatar"`
	ReceiveId  string `json:"receive_id"`
	FileSize   string `json:"file_size"`
	FileType   string `json:"file_type"`
	FileName   string `json:"file_name"`
	AVdata     string `json:"av_data"`
	MessageType int32 `json:"message_type"`
}


type MessageRequest struct {
	MessageType int32  `json:"messageType"` // 消息类型，1.单聊 2.群聊
	Uuid        string `json:"uuid"`        // 当前用户uuid
	FriendUuid  string `json:"friendUuid"`  // 好友用户uuid(单聊) 或 群聊uuid(群聊)
	Page        int32  `json:"page"`        // 分页页码
	PageSize    int32  `json:"pageSize"`    // 分页每页数量
}
