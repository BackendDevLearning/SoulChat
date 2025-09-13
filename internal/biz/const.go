package biz

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"time"
)

type UserRegisterReply struct {
	Phone    string
	UserName string
	Bio      string
	Image    string
	Token    string
}

type UserLoginReply struct {
	Phone    string
	UserName string
	Bio      string
	Image    string
	Token    string
}

type UserProfileReply struct {
	UserID uint32
	Tags   string

	FollowCount       uint32
	FanCount          uint32
	ViewCount         uint32
	NoteCount         uint32
	ReceivedLikeCount uint32
	CollectedCount    uint32
	CommentCount      uint32

	LastLoginIP string
	LastActive  *time.Time
	Status      string
}

type UserFollowFanReply struct {
	SelfID      uint32
	FollowCount uint32
	TargetID    uint32
	FanCount    uint32
}

type UserRelationshipReply struct {
	IsFollowing  bool
	IsFollowedBy bool
	IsMutual     bool
	IsBlocked    bool
	IsBlockedBy  bool
	IsFriend     bool
}

// IsValidPhone 校验手机号是否符合规则
func IsValidPhone(phone string) bool {
	// 中国大陆手机号规则：以 1 开头，第二位是 3-9，后面 9 位数字，总长度 11 位
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return reg.MatchString(phone)
}

// 密码映射为hash存储到数据库
func hashPassword(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// 验证密码
func verifyPassword(hashed, input string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input)); err != nil {
		return false
	}
	return true
}

func Contains(source []string, tg string) bool {
	for _, s := range source {
		if s == tg {
			return true
		}
	}
	return false
}

func Md5String(s string) string {
	//MD5 哈希器
	h := md5.New()
	//转换为字节数组写入哈希器
	h.Write([]byte(s))
	//h.Sum(nil) 表示返回哈希值
	//这个字节切片转换为 十六进制字符串
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

//func GenerateSession(uname string) string {
//    return Md5String(fmt.Sprintf("%s:%d", uname, rand.Intn(999999)))
//}

func GenerateSession(userName string) string {
	return Md5String(fmt.Sprintf("%s:%s", userName, "session"))
}
