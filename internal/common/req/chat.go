package req

// ChatMessageRequest 聊天消息请求结构
type ChatMessageRequest struct {
	SessionId     string `json:"session_id"`
	SendId        string `json:"send_id"`
	SendName      string `json:"send_name"`
	SendAvatar    string `json:"send_avatar"`
	ReceiveId     string `json:"receive_id"`
	ReceiveAvatar string `json:"receive_avatar"`
	Type          int32  `json:"type"`
	MessageType   int32  `json:"message_type"`
	Content       string `json:"content"`
	Url           string `json:"url"`
	Pic           string `json:"pic"`
	FileType      string `json:"file_type"`
	FileName      string `json:"file_name"`
	FileSize      string `json:"file_size"`
	AVdata        string `json:"av_data"`
}

type MessageRequest struct {
	MessageType int32  `json:"messageType"` // 消息类型，1.单聊 2.群聊
	Uuid        string `json:"uuid"`        // 当前用户uuid
	FriendUuid  string `json:"friendUuid"`  // 好友用户uuid(单聊) 或 群聊uuid(群聊)
	Page        int32  `json:"page"`        // 分页页码
	PageSize    int32  `json:"pageSize"`    // 分页每页数量
}
