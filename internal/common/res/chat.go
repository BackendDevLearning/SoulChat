package res

// GetMessageListRespond 消息列表响应结构
type GetMessageListRespond struct {
	SessionId     string `json:"sessionId"`
	SendId        string `json:"send_id"`
	SendName      string `json:"send_name"`
	SendAvatar    string `json:"send_avatar"`
	ReceiveId     string `json:"receive_id"`
	ReceiveAvatar string `json:"receive_avatar"`
	Type          int32  `json:"type"`
	MessageType   int32  `json:"messageType"`
	Content       string `json:"content"`
	Url           string `json:"url"`
	Pic					  string `json:"pic"`
	FileType      string `json:"file_type"`
	FileName      string `json:"file_name"`
	FileSize      string `json:"file_size"`
	AVdata        string `json:"av_data"`
	CreatedAt     string `json:"createdAt"`
}

// GetGroupMessageListRespond 群消息列表响应结构
type GetGroupMessageListRespond struct {
	SessionId     string `json:"sessionId"`
	SendId        string `json:"sendId"`
	SendName      string `json:"sendName"`
	SendAvatar    string `json:"sendAvatar"`
	ReceiveId     string `json:"receiveId"`
	ReceiveAvatar string `json:"receive_avatar"`
	Type          int32  `json:"type"`
	MessageType   int32  `json:"messageType"`
	Content       string `json:"content"`
	Url           string `json:"url"`
	Pic					  string `json:"pic"`
	FileType      string `json:"fileType"`
	FileName      string `json:"fileName"`
	FileSize      string `json:"fileSize"`
	AVdata        string `json:"av_data"`
	CreatedAt     string `json:"createdAt"`
}

// AVMessageRespond 音视频消息响应结构
type AVMessageRespond struct {
	SessionId     string `json:"sessionId"`
	SendId        string `json:"sendId"`
	SendName      string `json:"sendName"`
	SendAvatar    string `json:"sendAvatar"`
	ReceiveId     string `json:"receiveId"`
	ReceiveAvatar string `json:"receive_avatar"`
	Type          int32  `json:"type"`
	MessageType   int32  `json:"messageType"`
	Content       string `json:"content"`
	Url           string `json:"url"`
	Pic					  string `json:"pic"`
	FileType      string `json:"fileType"`
	FileName      string `json:"fileName"`
	FileSize      string `json:"fileSize"`
	AVdata        string `json:"avdata"`
	CreatedAt     string `json:"createdAt"`
}
