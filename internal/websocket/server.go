package websocket

import (
	"encoding/base64"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
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
		Broadcast:  make(chan []byte, 500),
		Register:   make(chan *Client, 50),
		Unregister: make(chan *Client, 50),
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

// Start TODO 如果多端（app/微信小程序/网页）同时在线，可能需要考虑互斥锁的问题，目前先不用
func (s *Server) Start() {
	for {
		select {
		case conn := <-s.Register:
			s.Clients[conn.Name] = conn
			msg := &v1.Message{
				From:    "System",
				To:      conn.Name,
				Content: "welcome!",
			}
			protoMsg, _ := proto.Marshal(msg)
			conn.Send <- protoMsg

		case conn := <-s.Unregister:
			if _, ok := s.Clients[conn.Name]; ok {
				close(conn.Send)
				delete(s.Clients, conn.Name)
			}

		case message := <-s.Broadcast:
			msg := &v1.Message{}
			err := proto.Unmarshal(message, msg)
			if err != nil {
				// ignore invalid payloads
				continue
			}

			if msg.To != "" {
				if msg.ContentType >= common.TEXT && msg.ContentType <= common.VIDEO {
					// 1.文字 2.普通文件 3.图片 4.音频 5.视频
					_, exits := s.Clients[msg.From]
					if exits {
						s.saveMessage(msg)
					}

					if msg.ContentType == common.MESSAGE_TYPE_USER {
						// 单聊
						client, ok := s.Clients[msg.To]
						if ok {
							msgByte, err := proto.Marshal(msg)
							if err == nil {
								client.Send <- msgByte
							}
						}
					} else if msg.MessageType == common.MESSAGE_TYPE_GROUP {
						// 群聊
						sendGroupMessage(msg, s)
					} else {
						clent, ok := s.Clients[msg.To]
						if ok {
							clent.Send <- message
						}
					}
				} else {
					// 6.语音聊天 7.视频聊天
					for id, conn := range s.Clients {
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

// 保存消息
// 主要实现文件上传到文件服务器 + 消息存储到数据库
// TODO 目前暂时将文件保存到本地静态目录下，后续考虑单独起文件服务器
func (s *Server) saveMessage(message *v1.Message) {
	if message.ContentType == 2 {
		// 普通的文件二进制上传
		message = SaveFile(message)
	} else if message.ContentType == 3 {
		// 保存图片
		message = SaveImg(message)
	}

	// 消息数据持久化到数据库
	msg := ConvertToMessage(message)
	err := s.mc.SaveMessage(msg)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func SaveFile(message *v1.Message) *v1.Message {
	dataBuffer, fileSuffix := ProcessBytes(message)

	contentType := util.GetContentTypeBySuffix(fileSuffix)
	url := uuid.New().String() + "." + fileSuffix

	// 保存到静态目录
	err := os.MkdirAll(staticBaseDir, 0755)
	if err != nil {
		log.Debug("创建目录失败:", err)
		return nil
	}
	path := filepath.Join(staticBaseDir, url)
	if err := os.WriteFile(path, dataBuffer, 0644); err != nil {
		log.Debug("写入文件失败:", err)
		return nil
	}

	// 修改消息信息
	message.ContentType = uint32(contentType)
	message.Url = url
	message.File = nil

	return message
}

func SaveImg(message *v1.Message) *v1.Message {
	var dataBuffer []byte
	var fileSuffix string

	if message.File != nil && len(message.File) > 0 {
		// 图片直接是二进制数据流
		dataBuffer, fileSuffix = ProcessBytes(message)
	} else if strings.HasPrefix(message.Content, "data:") {
		// 图片以base64形式传递过来
		dataBuffer, fileSuffix = ProcessBase64(message)
	} else {
		log.Debug("Content 既不是文件二进制，也不是 base64，无法处理")
		return nil
	}

	url := uuid.New().String() + "." + fileSuffix

	// 保存到静态目录
	err := os.MkdirAll(staticBaseDir, 0755)
	if err != nil {
		log.Debug("创建目录失败:", err)
		return nil
	}
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

func ProcessBytes(message *v1.Message) (dataBuffer []byte, fileSuffix string) {
	fileSuffix = util.GetFileType(message.File)
	if fileSuffix == "" {
		fileSuffix = strings.ToLower(message.FileSuffix)
	}

	dataBuffer = message.File
	return dataBuffer, fileSuffix
}

// base64编码数据格式："data:image/png;base64,content"
func ProcessBase64(message *v1.Message) (dataBuffer []byte, fileSuffix string) {
	commaIndex := strings.Index(message.Content, ",")
	if commaIndex < 0 {
		log.Debug("base64 数据格式错误")
		return nil, ""
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

	dataBuffer, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Debug("base64 解码失败:", err)
		return nil, ""
	}

	return dataBuffer, fileSuffix
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
