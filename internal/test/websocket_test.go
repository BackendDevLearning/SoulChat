package test

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strings"
	"time"

	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"kratos-realworld/api/conduit/v1"
	"log"
	"testing"
)

func TestWebSocket(t *testing.T) {
	// 1. 建立 WebSocket 连接
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjk3MTE3OTIwMCwidXNlcmlkIjo2fQ.7uGIGpfq6OUspnsVT3FlSNX3iUhPuXkwjLDwkC_5Cl0"
	header := http.Header{}
	header.Add("Authorization", fmt.Sprintf("Token %s", tokenString))

	// 两种方式获取WebSocket服务的token
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws", header)
	//conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:8000/ws?token=%s", tokenString), nil)

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
	ready := make(chan struct{})

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("ReadMessage 错误:", err)
				return
			}
			log.Println("收到服务器消息:", string(msg))

			// 收到注册确认信号后，通知主线程可以发消息
			if strings.Contains(string(msg), "welcome!") {
				close(ready)
			}
		}
	}()

	// 4. 模拟前端构造proto消息，并转为二进制
	//msg := CreateTextMessage()
	msg := CreateFileMessage()
	//msg := CreateImageMessage()

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshal proto:", err)
	}

	// 5. 发送一次业务消息
	select {
	case <-ready:
		log.Println("收到ready信号，client在服务器注册完成，开始发送业务消息！")
	case <-time.After(5 * time.Second):
		log.Println("注册超时！")
	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Println("发送业务消息失败:", err)
	} else {
		log.Println("发送业务消息成功")
	}

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

	select {}
}

// 1. 生成文本消息
func CreateTextMessage() *v1.Message {
	return &v1.Message{
		Avatar:       "1",
		FromUserName: "2",
		From:         "6",
		To:           "5",
		Content:      "hello world!",
		MessageType:  1,
		ContentType:  1,
		Type:         "type",
		Url:          "",
		FileSuffix:   "",
		File:         nil,
	}
}

// 2. 生成文件消息
func CreateFileMessage() *v1.Message {
	filePath := "data/attention is all you need.pdf"

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}

	return &v1.Message{
		Avatar:       "1",
		FromUserName: "2",
		From:         "6",
		To:           "5",
		Content:      "",
		MessageType:  1,
		ContentType:  2,
		Type:         "type",
		Url:          "http://localhost:3000",
		FileSuffix:   "pdf",
		File:         fileBytes,
	}
}

// 3. 生成图片消息
func CreateImageMessage() *v1.Message {
	// Base64编码
	//content := "data:image/png;base64," + data.GetImgBase64()
	//return &v1.Message{
	//	Avatar:       "1",
	//	FromUserName: "2",
	//	From:         "6",
	//	To:           "5",
	//	Content:      content,
	//	MessageType:  1,
	//	ContentType:  3,
	//	Type:         "type",
	//	Url:          "http://localhost:3000",
	//	FileSuffix:   "",
	//	File:         nil,
	//}

	// 文件字节流
	//filePath := "data/sky.jpg"
	filePath := "data/sky_compress.jpg"

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}

	return &v1.Message{
		Avatar:       "1",
		FromUserName: "2",
		From:         "6",
		To:           "5",
		Content:      "",
		MessageType:  1,
		ContentType:  3,
		Type:         "type",
		Url:          "http://localhost:3000",
		FileSuffix:   "jpg",
		File:         fileBytes,
	}
}
