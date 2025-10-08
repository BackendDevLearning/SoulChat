package test

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"kratos-realworld/api/conduit/v1"
	"log"
	"testing"
)

func TestWebSocket(t *testing.T) {

	// 测试连接
	conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8000/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjk3MTE3OTIwMCwidXNlcmlkIjo2fQ.7uGIGpfq6OUspnsVT3FlSNX3iUhPuXkwjLDwkC_5Cl0", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	// 模拟用户输入 JSON
	jsonMessage := `{
		"avatar":       "1",
		"fromUserName": "2",
		"from":         "2935",
		"to":           "2661",
		"content":      "hello world",
		"messageType":  1,
		"contentType":  1,
		"type":         "type",
		"url":          "http://localhost:3000",
		"fileSuffix":   ".md",
		"file":         "dGVzdGRhdGE="
	}`

	// 解析 JSON -> proto.Message
	var msg v1.Message
	if err := json.Unmarshal([]byte(jsonMessage), &msg); err != nil {
		log.Fatal("unmarshal json:", err)
	}
	fmt.Println("msg:", msg)

	// 序列化成二进制
	data, err := proto.Marshal(&msg)
	fmt.Println("data:", data)

	if err != nil {
		log.Fatal("marshal proto:", err)
	}

	fmt.Println("data", string(data))

	// 发送二进制消息
	err = conn.WriteMessage(websocket.BinaryMessage, data)
	//err = conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}

	// 读取服务端返回
	//msgType, reply, err := conn.ReadMessage()
	//if err != nil {
	//	log.Println("read:", err)
	//	return
	//}
	//
	//log.Printf("server reply type=%v, raw bytes=%v", msgType, reply)
	//
	//// 尝试反序列化为 proto.Message
	//var replyMsg v1.Message
	//if err := proto.Unmarshal(reply, &replyMsg); err == nil {
	//	log.Printf("server reply (proto): %+v", replyMsg)
	//} else {
	//	log.Printf("server reply (string): %s", string(reply))
	//}
}
