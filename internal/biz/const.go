package biz

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// 对data结构
// data 存，data取的结构
type UserRegisterTB struct {
	Phone    string
	Username string
	Password string
}

type UserRegisterReply struct {
	Phone    string
	Username string
	Token    string
	Bio      string
	Image    string
}

// biz的一些公共函数
func hashPassword(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

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

func GenerateSession(userName string) string {
	return Md5String(fmt.Sprintf("%s:%s", userName, "session"))
}
