package service

import (
	stdErrors "errors"
	v1 "kratos-realworld/api/conduit/v1"

	"github.com/go-kratos/kratos/v2/errors"
)

// ErrorToRes biz层错误返回给前端更直观的res结构
func ErrorToRes(err error) *v1.Res {
	if err == nil {
		return &v1.Res{
			Code:   200,
			Reason: "OK",
			Msg:    "success",
		}
	}

	var e *errors.Error
	if stdErrors.As(err, &e) {
		return &v1.Res{
			Code:   e.Code,
			Reason: e.Reason,
			Msg:    e.Message,
		}
	}

	// 普通 error
	return &v1.Res{
		Code:   500,
		Reason: "UNKNOWN",
		Msg:    err.Error(),
	}
}

const (
	ReqUuid          = "uuid"
	UserInfoPrefix   = "userinfo_"
	SessionKeyPrefix = "session_"
)

const (
	GenderMale   = "male"
	GenderFeMale = "female"
)

const (
	SessionKey   = "user_session"
	CookieExpire = 3600
)
