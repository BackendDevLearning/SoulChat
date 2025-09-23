package websocket

import (
	"github.com/go-kratos/kratos/v2/log"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"sync"
	"github.com/google/uuid"
	"github.com/gogo/protobuf/proto"
	"kratos-realworld/api/conduit/v1"
	"kratos-realworld/common"
)

var MyServer = NewServer()

type Server struct {
	Clients   map[string]*Client
	mutex     *sync.Mutex
	Broadcast chan []byte
	Register  chan *Client
	Ungister  chan *Client
}

func NewServer() *Server {
	return &Server{
		mutex:     &sync.Mutex{},
		Clients:   make(map[string]*Client),
		Broadcast: make(chan []byte),
		Register:  make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func ConsumerKafkaMsg(data []byte) {
	MyServer.Broadcast <- data
}

func (s *Server) Start() {
	for {
		select {
		case conn := <- s.Register:
			s.Clients[conn.Name] = conn
			msg := &v1.Message{
				From:    "System",
				To:      conn.Name,
				Content: "welcome!",
			}
			protoMsg, _ := proto.Marshal(msg)
			conn.Send <- protoMsg

		case conn := <- s.Ungister:
			if _, ok := s.Clients[conn.Name]; ok {
				close(conn.Send)
				delete(s.Clients, conn.Name)
			}

		case message := <- s.Broadcast:
			msg := &v1.Message{}
			err := proto.Unmarshal(message, msg)
			if err != nil {
				log.Errorf("failed to unmarshal message: %v", err)
				continue
			}
			
			if msg.To == "" {
				if msg.ContentType >= common.TEXT && msg.ContentType <= common.VIDEO {
					_, exits := s.Clients[msg.From]
					if exits {
						saveMessage(msg)
					}

					if msg.Message == common.MESSAGE_TYPE_USER {
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(msg)
							if err == nil {
								client.Send <- msgByte
							}
						}
						else if msg.MessageType == common.MESSAGE_TYPE_GROUP {
							sendGroupMessage(msg, s)
						}
					} else {
						clent, ok := s.Clients[msg.To]
						if ok {
							clent.Send <- message
						}
					}
				} else {
					for id, conn := s.Clients {
						select {
						case conn.Send <- message:
						default:
							close(conn.Send)
							delete(s.Clients, id)
						}
					}
				}
			}
		}
	}
}