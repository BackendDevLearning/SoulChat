package biz

import (
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
