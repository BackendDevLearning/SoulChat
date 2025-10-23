package test

import (
	"testing"
	"time"
)

// 启动两个 WebSocket 客户端（A 和 B）
func TestWebSocketAB(t *testing.T) {
	go TestWebSocketA(t) // 启动客户端 A

	go TestWebSocketB(t) // 启动客户端 B

	// 阻塞主 goroutine，防止程序退出
	for {
		time.Sleep(5 * time.Second)
	}
}
