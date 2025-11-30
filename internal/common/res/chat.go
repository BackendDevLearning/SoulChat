package res

// GetMessageListRespond 消息列表响应结构
type GetMessageListRespond struct {
	Type       int8   `json:"type"`
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
}

// GetGroupMessageListRespond 群消息列表响应结构
type GetGroupMessageListRespond struct {
	SessionId  string `json:"sessionId"`
	SendId     string `json:"sendId"`
	SendName   string `json:"sendName"`
	SendAvatar string `json:"sendAvatar"`
	ReceiveId  string `json:"receiveId"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	FileSize   string `json:"fileSize"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	CreatedAt  string `json:"createdAt"`
	MessageType int8 `json:"messageType"`
}

// AVMessageRespond 音视频消息响应结构
type AVMessageRespond struct {
	SendId     string `json:"sendId"`
	SendName   string `json:"sendName"`
	SendAvatar string `json:"sendAvatar"`
	ReceiveId  string `json:"receiveId"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Url        string `json:"url"`
	FileSize   string `json:"fileSize"`
	FileName   string `json:"fileName"`
	FileType   string `json:"fileType"`
	CreatedAt  string `json:"createdAt"`
	AVdata     string `json:"avdata"`
	MessageType int8 `json:"messageType"`
}
