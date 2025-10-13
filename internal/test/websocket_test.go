package test

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"

	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"kratos-realworld/api/conduit/v1"
	"log"
	"testing"
)

func TestWebSocket(t *testing.T) {
	// 1. 建立 WebSocket 连接
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjk3MTE3OTIwMCwidXNlcmlkIjo2fQ.7uGIGpfq6OUspnsVT3FlSNX3iUhPuXkwjLDwkC_5Cl0", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// 2. 处理服务器 Ping，回复 Pong
	conn.SetPingHandler(func(appData string) error {
		log.Println("收到服务器 Ping，回复 Pong")
		return conn.WriteMessage(websocket.PongMessage, nil)
	})

	// 3. 循环读取服务器消息
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("ReadMessage 错误:", err)
				return
			}
			log.Println("收到服务器消息:", string(msg))
		}
	}()

	// 4. 准备要发送的 JSON 消息
	// 1: 文字
	//jsonMessage := `{
	//	"avatar":       "1",
	//	"fromUserName": "2",
	//	"from":         "6",
	//	"to":           "5",
	//	"content":      "hello world",
	//	"messageType":  1,
	//	"contentType":  1,
	//	"type":         "type",
	//	"url":          "",
	//	"fileSuffix":   "",
	//	"file":         ""
	//}`

	// 2: 普通文件
	//jsonMessage := `{
	//	"avatar":       "1",
	//	"fromUserName": "2",
	//	"from":         "6",
	//	"to":           "5",
	//	"content":      "",
	//	"messageType":  1,
	//	"contentType":  2,
	//	"type":         "type",
	//	"url":          "http://localhost:3000",
	//	"fileSuffix":   ".md",
	//	"file":         "dGVzdGRhdGE="
	//}`

	// 3: 图片 base64
	jsonMessage := fmt.Sprintf(`{
		"avatar":       "1",
		"fromUserName": "2",
		"from":         "6",
		"to":           "5",
		"content":      "data:image/png;base64,%s",
		"messageType":  1,
		"contentType":  3,
		"type":         "type",
		"url":          "http://localhost:3000",
		"fileSuffix":   "png",
		"file":         ""
	}`, imgBase64)

	//jsonMessage := `{
	//	"avatar":       "1",
	//	"fromUserName": "2",
	//	"from":         "6",
	//	"to":           "5",
	//	"content":      "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8Xw8AAoMBgQFG4RcAAAAASUVORK5CYII=",
	//	"messageType":  1,
	//	"contentType":  3,
	//	"type":         "type",
	//	"url":          "http://localhost:3000",
	//	"fileSuffix":   "png",
	//	"file":         ""
	//}`

	// JSON → Go struct
	var msg v1.Message
	if err := json.Unmarshal([]byte(jsonMessage), &msg); err != nil {
		log.Fatal("unmarshal json:", err)
	}
	fmt.Println("msg:", msg)

	// Go struct → protobuf 二进制
	data, err := proto.Marshal(&msg)
	if err != nil {
		log.Fatal("marshal proto:", err)
	}
	fmt.Println("data:", data)

	// 5. 发送一次业务消息
	time.Sleep(2000 * time.Millisecond) // 给服务器注册完成
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Println("发送业务消息失败:", err)
	} else {
		log.Println("发送业务消息成功")
	}

	//time.Sleep(100 * time.Millisecond)

	// 6. 循环发送消息（可按需加心跳或业务消息）
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("心跳发送失败:", err)
				return
			}
			log.Println("心跳发送成功")
		}
	}()

	// 7. 阻塞主 goroutine，保持连接
	// 运行这个测试脚本时，这个测试脚本相当于一个主协程，内部其它协程都算子协程，他们依赖主协程不退出，否则程序会结束，直接defer conn.Close()
	select {}
}
