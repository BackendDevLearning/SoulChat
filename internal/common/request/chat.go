package request

// ChatMessageRequest 聊天消息请求结构
type ChatMessageRequest struct {
	SessionId string `json:"sessionId"`
	Type      string `json:"type"`      // 消息类型：Text, File, AudioOrVideo
	Content   string `json:"content"`   // 文本内容
	Url       string `json:"url"`       // 文件URL
	SendId    string `json:"sendId"`    // 发送者ID
	SendName  string `json:"sendName"`  // 发送者名称
	ReceiveId string `json:"receiveId"` // 接收者ID
	FileSize  string `json:"fileSize"`  // 文件大小
	FileType  string `json:"fileType"`  // 文件类型
	FileName  string `json:"fileName"`  // 文件名
	AVdata    string `json:"avdata"`    // 音视频数据
}

type AVData struct {
	MessageId string `json:"messageId"`
	Type      string `json:"type"`
}

type MessageRequest struct {
	Type      int    `json:"type"`
	Content   string `json:"content"`
	Url       string `json:"url"`
	SendId    string `json:"send_id"`
	ReceiveId string `json:"receive_id"`
	FileType  string `json:"file_type"`
	FileName  string `json:"file_name"`
	FileSize  int    `json:"file_size"`
}



