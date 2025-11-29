package common

// 消息类型常量
const (
	MessageTypeText         = 0
	MessageTypeAudio        = 1
	MessageTypeFile         = 2
	MessageTypeVideo 		= 3
)

// 消息状态常量
const (
	MessageStatusUnsent = 0
	MessageStatusSent   = 1
)

// 常量定义
const (
	REDIS_TIMEOUT = 30 // Redis 缓存超时时间（分钟）
)
