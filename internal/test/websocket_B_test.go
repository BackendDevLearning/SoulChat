package test

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"time"

	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"kratos-realworld/api/conduit/v1"
	"log"
	"testing"
)

func TestWebSocketB(t *testing.T) {
	// 1. 建立 WebSocket 连接
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjk3MTE3OTIwMCwidXNlcmlkIjo1fQ.IYIxodOC_-Lc4LczpvfiSxB39f1NTSRNx4x0CC4C2kI"
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Token %s", tokenString))
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", header)

	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// 2. 处理服务器 Ping，回复 Pong
	conn.SetPingHandler(func(appData string) error {
		log.Println("B 收到服务器 Ping，回复 Pong")
		return conn.WriteMessage(websocket.PongMessage, nil)
	})

	// 3. 循环读取服务器消息
	ready := make(chan struct{})

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("B ReadMessage 错误:", err)
				return
			}
			log.Println("B 收到服务器消息:", string(msg))

			// 收到注册确认信号后，通知主线程可以发消息
			if strings.Contains(string(msg), "welcome!") {
				close(ready)
			}
		}
	}()

	// 4. 准备要发送的 JSON 消息
	jsonMessage := `{
		"avatar":       "1",
		"fromUserName": "2",
		"from":         "5",
		"to":           "6",
		"content":      "My name is B!",
		"messageType":  1,
		"contentType":  1,
		"type":         "type",
		"url":          "",
		"fileSuffix":   "",
		"file":         ""
	}`

	// JSON → Go struct
	var msg v1.Message
	if err := json.Unmarshal([]byte(jsonMessage), &msg); err != nil {
		log.Fatal("unmarshal json:", err)
	}

	// Go struct → protobuf 二进制
	data, err := proto.Marshal(&msg)
	if err != nil {
		log.Fatal("marshal proto:", err)
	}

	// 5. 发送一次业务消息
	select {
	case <-ready:
		log.Println("B 收到ready信号，client在服务器注册完成，开始发送业务消息！")
	case <-time.After(5 * time.Second):
		log.Println("B 注册超时！")
	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Println("B 发送业务消息失败:", err)
	} else {
		log.Println("B 发送业务消息成功")
	}

	// 6. 循环发送心跳消息
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("B 心跳发送失败:", err)
				return
			}
			log.Println("B 心跳发送成功")
		}
	}()

	select {}
}
