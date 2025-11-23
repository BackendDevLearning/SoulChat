package chat

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// normalizePath 去除路径中/static之前的所有内容，防止ip前缀引入
func normalizePath(path string) string {
	if path == "" {
		return path
	}
	idx := strings.Index(path, "/static")
	if idx >= 0 {
		return path[idx:]
	}
	return path
}

// GetNowAndLenRandomString 生成指定长度的随机字符串（基于当前时间戳）
func GetNowAndLenRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateMessageUUID 生成消息UUID
func GenerateMessageUUID() string {
	return fmt.Sprintf("M%s", GetNowAndLenRandomString(11))
}

