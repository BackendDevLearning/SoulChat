package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/kafka"
)

type Client struct {
	Conn *websocket.Conn
	Name string
	Send chan []byte
}

func (c *Client) Read() {
	defer func() {
		MyServer.Unregister <- c
		c.Conn.Close()
	}()
	for {
		c.Conn.PongHandler()
		_, message, err := c.Conn.ReadMessage()

		fmt.Println("message:", message)

		if err != nil {
			MyServer.Unregister <- c
			c.Conn.Close()
			break
		}
		msg := &v1.Message{}
		proto.Unmarshal(message, msg)

		if msg.Type == common.HEAT_BEAT {
			pong := &v1.Message{
				Content: common.PONG,
				Type:    common.HEAT_BEAT,
			}
			pongByte, err2 := proto.Marshal(pong)
			if err2 == nil {
				c.Conn.WriteMessage(websocket.BinaryMessage, pongByte)
			}
		} else {
			// 直接放到消息队列里面，回调放到broadcast里面
			kafka.Send(message)
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		c.Conn.WriteMessage(websocket.BinaryMessage, message)
	}
}
