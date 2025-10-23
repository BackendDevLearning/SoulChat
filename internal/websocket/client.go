package websocket

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
	"sync"
	"time"

	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/kafka"
)

type Client struct {
	Conn      *websocket.Conn
	Name      string
	Send      chan []byte
	closeOnce sync.Once
}

const (
	pongWait   = 60 * time.Second // 服务器等待客户端 pong 的最大时间
	pingPeriod = 50 * time.Second // 服务器主动发送 ping 的周期，通常 < pongWait
)

func (c *Client) Read() {
	defer func() {
		MyServer.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error {
		log.Debug("收到客户端 Pong")
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		//c.Conn.PongHandler()
		_, message, err := c.Conn.ReadMessage()

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
				//c.Conn.WriteMessage(websocket.BinaryMessage, pongByte)
				c.Send <- pongByte // 发给写协程
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

	// 定时发送心跳包维持连接
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			// 监听 c.Send 通道中是否有要发送的消息
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// channel 关闭，发送 close 消息
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.Conn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				return
			}

		case <-ticker.C:
			// 定期发送 Ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
				return
			}
		}
	}
}
