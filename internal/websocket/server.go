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
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/biz"
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

// 发送给群组消息,需要查询该群所有人员依次发送
func sendGroupMessage(msg *protocol.Message, s *Server) {
	// 发送给群组的消息，查找该群所有的用户进行发送
	users := service.GroupService.GetUserIdByGroupUuid(msg.To)
	for _, user := range users {
		if user.Uuid == msg.From {
			continue
		}

		client, ok := s.Clients[user.Uuid]
		if !ok {
			continue
		}

		fromUserDetails := service.UserService.GetUserDetails(msg.From)
		// 由于发送群聊时，from是个人，to是群聊uuid。所以在返回消息时，将form修改为群聊uuid，和单聊进行统一
		msgSend := protocol.Message{
			Avatar:       fromUserDetails.Avatar,
			FromUsername: msg.FromUsername,
			From:         msg.To,
			To:           msg.From,
			Content:      msg.Content,
			ContentType:  msg.ContentType,
			Type:         msg.Type,
			MessageType:  msg.MessageType,
			Url:          msg.Url,
		}

		msgByte, err := proto.Marshal(&msgSend)
		if err == nil {
			client.Send <- msgByte
		}
	}
}

// 保存消息，如果是文本消息直接保存，如果是文件，语音等消息，保存文件后，保存对应的文件路径
func saveMessage(message *v1.Message) {
	// 如果上传的是base64字符串文件，解析文件保存
	if message.ContentType == 2 {
		url := uuid.New().String() + ".png"
		index := strings.Index(message.Content, "base64")
		index += 7

		content := message.Content
		content = content[index:]

		dataBuffer, dataErr := base64.StdEncoding.DecodeString(content)
		if dataErr != nil {
			log.Logger.Error("transfer base64 to file error", log.String("transfer base64 to file error", dataErr.Error()))
			return
		}
		err := ioutil.WriteFile(config.GetConfig().StaticPath.FilePath+url, dataBuffer, 0666)
		if err != nil {
			log.Logger.Error("write file error", log.String("write file error", err.Error()))
			return
		}
		message.Url = url
		message.Content = ""
	} else if message.ContentType == 3 {
		// 普通的文件二进制上传
		fileSuffix := util.GetFileType(message.File)
		nullStr := ""
		if nullStr == fileSuffix {
			fileSuffix = strings.ToLower(message.FileSuffix)
		}
		contentType := util.GetContentTypeBySuffix(fileSuffix)
		url := uuid.New().String() + "." + fileSuffix
		err := ioutil.WriteFile(config.GetConfig().StaticPath.FilePath+url, message.File, 0666)
		if err != nil {
			log.Logger.Error("write file error", log.String("write file error", err.Error()))
			return
		}
		message.Url = url
		message.File = nil
		message.ContentType = contentType
	}

	biz.SaveMessage(*message)
}
