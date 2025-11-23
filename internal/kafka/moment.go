package kafka

import "encoding/json"

type MomentNotifyHandler struct {
	appPush string
}

func (h *MomentNotifyHandler) Topic() string {
	return MOMENT_TOPIC
}

func (h *MomentNotifyHandler) Push(msg []byte) error {
	var moment MomentMessage
	if err := json.Unmarshal(msg, &moment); err != nil {
		return err
	}
	//todo
	return nil
}
