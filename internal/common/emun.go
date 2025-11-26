package common

// 消息类型常量
const (
	MessageTypeText         = "Text"
	MessageTypeFile         = "File"
	MessageTypeAudioOrVideo = "AudioOrVideo"
)

// 消息状态常量
const (
	MessageStatusUnsent = "Unsent"
	MessageStatusSent   = "Sent"
)

// 常量定义
const (
	REDIS_TIMEOUT = 30 // Redis 缓存超时时间（分钟）
)
