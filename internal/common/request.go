package common

type MessageRequest struct {
	MessageType int32  `json:"messageType"` // 消息类型，1.单聊 2.群聊
	Uuid        string `json:"uuid"`        // 当前用户uuid
	FriendUuid  string `json:"friendUuid"`  // 好友用户uuid(单聊) 或 群聊uuid(群聊)
	Page        int32  `json:"page"`        // 分页页码
	PageSize    int32  `json:"pageSize"`    // 分页每页数量
}
