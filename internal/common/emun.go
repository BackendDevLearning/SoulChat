package common

// 消息类型常量
const (
	MessageTypeText         			= 0
	MessageTypeVoice        			= 1
	MessageTypeFile        				= 2
	MessageTypeAudioOrVideo 			= 3
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

// 聊天类型常量
const (
	MESSAGE_TYPE_SINGLE = 1
	MESSAGE_TYPE_GROUP  = 2
)

const (
	DefaultCacheTTL = 24 * time.Hour
	GroupMyGroupListCacheTTL = 12 * time.Hour
)

const (
	GroupAdmin = 1
	GroupMember = 0
)