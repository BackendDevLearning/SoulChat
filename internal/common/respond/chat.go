package respond

// GetMessageListRespond 单个用户消息列表响应结构
type GetMessageListRespond struct {
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
}

// GetGroupMessageListRespond 群组消息列表响应结构
type GetGroupMessageListRespond struct {
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
}

// AVData 音视频数据
type AVData struct {
	MessageId string `json:"messageId"`
	Type      string `json:"type"`
}

