package websocket

import (
	"encoding/base64"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"io/ioutil"
	"kratos-realworld/internal/biz"
	"kratos-realworld/internal/pkg/util"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	//"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/proto"
	v1 "kratos-realworld/api/conduit/v1"
	bizChat "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/common"
	// "kratos-realworld/internal/pkg/util"
	"fmt"
	// bizUser "kratos-realworld/internal/biz/user"
)

var MyServer *Server

func InitWebsocketServer(mc *biz.MessageUseCase) {
	MyServer = NewServer(mc)
}

type Server struct {
	Clients    map[string]*Client
	mutex      *sync.Mutex
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mc         *biz.MessageUseCase
}

func NewServer(mc *biz.MessageUseCase) *Server {
	return &Server{
		mutex:      &sync.Mutex{},
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		mc:         mc,
	}
}

// staticBaseDir 指定静态文件保存目录，由 main 在启动时设置
var staticBaseDir = "./static/"

func SetStaticBaseDir(dir string) {
	if dir == "" {
		return
	}
	staticBaseDir = dir
}

func ConsumerKafkaMsg(data []byte) {
	MyServer.Broadcast <- data
}

func (s *Server) Start() {
	for {
		select {
		case conn := <-s.Register:
			s.mutex.Lock()
			s.Clients[conn.Name] = conn
			s.mutex.Unlock()
			msg := &v1.Message{
				From:    "System",
				To:      conn.Name,
				Content: "welcome!",
			}
			protoMsg, _ := proto.Marshal(msg)
			conn.Send <- protoMsg

		case conn := <-s.Unregister:
			s.mutex.Lock()
			if _, ok := s.Clients[conn.Name]; ok {
				close(conn.Send)
				delete(s.Clients, conn.Name)
			}
			s.mutex.Unlock()

		case message := <-s.Broadcast:
			msg := &v1.Message{}
			err := proto.Unmarshal(message, msg)
			if err != nil {
				// ignore invalid payloads
				continue
			}

			if msg.To != "" {
				if msg.ContentType >= common.TEXT && msg.ContentType <= common.VIDEO {
					s.mutex.Lock()
					_, exits := s.Clients[msg.From]
					s.mutex.Unlock()
					if exits {
						s.saveMessage(msg)
					}

					if msg.ContentType == common.MESSAGE_TYPE_USER {
						s.mutex.Lock()
						client, ok := s.Clients[msg.To]
						s.mutex.Unlock()
						if ok {
							msgByte, err := proto.Marshal(msg)
							if err == nil {
								client.Send <- msgByte
							}
						}
					} else if msg.MessageType == common.MESSAGE_TYPE_GROUP {
						sendGroupMessage(msg, s)
					} else {
						s.mutex.Lock()
						clent, ok := s.Clients[msg.To]
						s.mutex.Unlock()
						if ok {
							clent.Send <- message
						}
					}
				} else {
					s.mutex.Lock()
					for id, conn := range s.Clients {
						select {
						case conn.Send <- message:
						default:
							close(conn.Send)
							delete(s.Clients, id)
						}
					}
					s.mutex.Unlock()
				}
			}
		}
	}
}

// 发送给群组消息,需要查询该群所有人员依次发送
func sendGroupMessage(msg *v1.Message, s *Server) {
	// 发送给群组的消息，查找该群所有的用户进行发送
	// todo: 实现接口
	// users := service.GroupService.GetUserIdByGroupUuid(msg.To)
	// for _, user := range users {
	// 	if user.Uuid == msg.From {
	// 		continue
	// 	}

	// 	client, ok := s.Clients[user.Uuid]
	// 	if !ok {
	// 		continue
	// 	}

	// 	fromUserDetails := service.UserService.GetUserDetails(msg.From)
	// 	// 由于发送群聊时，from是个人，to是群聊uuid。所以在返回消息时，将form修改为群聊uuid，和单聊进行统一
	// 	msgSend := v1.Message{
	// 		Avatar:       fromUserDetails.Avatar,  //todo: 实现接口
	// 		FromUsername: msg.FromUsername,
	// 		From:         msg.To,
	// 		To:           msg.From,
	// 		Content:      msg.Content,
	// 		ContentType:  msg.ContentType,
	//         Type:         msg.Type,
	//         MessageType:  msg.MessageType,
	// 		Url:          msg.Url,
	// 	}

	// 	msgByte, err := proto.Marshal(&msgSend)
	// 	if err == nil {
	// 		client.Send <- msgByte
	// 	}
	// }
	fmt.Println("sendGroupMessage")
}

// 保存消息，如果是文本消息直接保存，如果是文件，语音等消息，保存文件后，保存对应的文件路径
func (s *Server) saveMessage(message *v1.Message) {
	if message.ContentType == 2 {
		// 普通的文件二进制上传
		fileSuffix := util.GetFileType(message.File)
		nullStr := ""
		if nullStr == fileSuffix {
			fileSuffix = strings.ToLower(message.FileSuffix)
		}
		contentType := util.GetContentTypeBySuffix(fileSuffix)
		url := uuid.New().String() + "." + fileSuffix
		// 保存到静态目录
		path := filepath.Join(staticBaseDir, url)
		err := ioutil.WriteFile(path, message.File, 0666)
		if err != nil {
			return
		}
		message.Url = url
		message.File = nil
		message.ContentType = uint32(contentType)
	} else if message.ContentType == 3 {
		// 保存图片
		message = ProcessImg(message)
	}

	// 消息数据持久化到数据库
	msg := ConvertToMessage(message)
	err := s.mc.SaveMessage(msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func ProcessImg(message *v1.Message) *v1.Message {
	var dataBuffer []byte
	var fileSuffix string

	if message.File != nil && len(message.File) > 0 {
		// 图片直接是二进制数据流
		dataBuffer = message.File
		fileSuffix = strings.ToLower(message.FileSuffix)
		if fileSuffix == "" {
			fileSuffix = "bin" // 默认后缀
		}
	} else if strings.HasPrefix(message.Content, "data:") {
		// 图片以base64形式传递过来
		commaIndex := strings.Index(message.Content, ",")
		if commaIndex < 0 {
			log.Debug("base64 数据格式错误")
			return nil
		}

		header := message.Content[:commaIndex]
		content := message.Content[commaIndex+1:]

		// 自动获取文件类型
		if strings.Contains(header, "image/png") {
			fileSuffix = "png"
		} else if strings.Contains(header, "image/jpeg") {
			fileSuffix = "jpg"
		} else {
			fileSuffix = "bin" // 默认
		}

		buf, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			log.Debug("base64 解码失败:", err)
			return nil
		}
		dataBuffer = buf
	} else {
		log.Debug("Content 既不是文件二进制，也不是 base64，无法处理")
		return nil
	}

	// 保存到静态目录
	err := os.MkdirAll(staticBaseDir, 0755)
	if err != nil {
		log.Debug("创建目录失败")
		log.Debug(err)
		return nil
	}
	url := uuid.New().String() + "." + fileSuffix
	path := filepath.Join(staticBaseDir, url)
	if err := os.WriteFile(path, dataBuffer, 0644); err != nil {
		log.Debug("写入文件失败:", err)
		return nil
	}

	// 修改消息信息
	message.Url = url
	message.Content = ""
	message.File = nil

	return message
}

func ConvertToMessage(msg *v1.Message) *bizChat.MessageTB {
	if msg == nil {
		return nil
	}
	now := time.Now()

	return &bizChat.MessageTB{
		CreatedAt:   &now,
		UpdatedAt:   &now,
		FromUserID:  msg.From,
		ToUserID:    msg.To,
		Content:     msg.Content,
		MessageType: uint16(msg.MessageType),
		ContentType: uint16(msg.ContentType),
		Url:         msg.Url,
		Pic:         "",
	}
}
