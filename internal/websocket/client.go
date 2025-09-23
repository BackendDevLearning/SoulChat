package websocket

import (
	"github.com/go-kratos/kratos/v2/log"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"sync"
	"github.com/gogo/protobuf/proto"
)

type Client struct {
	Conn *websocket.Conn
	Name string
	Send chan []byte
}

func (c *Client) Read() {
	defer func() {
		MyServer.Ungister <- c
		c.Conn.Close()
	}()
	for {
		c.Conn.PongHandler()
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("failed to read message: %v", err)
			MyServer.Ungister <- c
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
			// 这里要修改
			if conf.GetConf().MsgChannelType.ChannelType == common.KAFKA {
				kafka.Send(message)
			} else {
				MyServer.Broadcast <- message
			}
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
