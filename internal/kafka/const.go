package kafka

type MomentMessage struct {
	UserID        uint32   `json:"user_id"`
	MomentID      uint32   `json:"moment_id"`
	Action        string   `json:"action"`
	ReceiveBoxIDs []uint32 `json:"receive_box_ids"`
}

const (
	// 动态相关
	MOMENT_TOPIC = "moment_topic"
)
