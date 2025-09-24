package common

import "time"

type ResponseMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
type MessageResponse struct {
	ID           int32     `json:"id" gorm:"primarykey"`
	FromUserId   int32     `json:"fromUserId" gorm:"index"`
	ToUserId     int32     `json:"toUserId" gorm:"index"`
	Content      string    `json:"content" gorm:"type:varchar(2500)"`
	ContentType  int16     `json:"contentType" gorm:"comment:'消息内容类型：1文字，2语音，3视频'"`
	CreatedAt    time.Time `json:"createAt"`
	FromUsername string    `json:"fromUsername"`
	ToUsername   string    `json:"toUsername"`
	Avatar       string    `json:"avatar"`
	Url          string    `json:"url"`
}

func SuccessMsg(data interface{}) *ResponseMsg {
	msg := &ResponseMsg{
		Code: 0,
		Msg:  "SUCCESS",
		Data: data,
	}
	return msg
}

func FailMsg(msg string) *ResponseMsg {
	msgObj := &ResponseMsg{
		Code: -1,
		Msg:  msg,
	}
	return msgObj
}

func FailCodeMsg(code int, msg string) *ResponseMsg {
	msgObj := &ResponseMsg{
		Code: code,
		Msg:  msg,
	}
	return msgObj
}
