package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// NewErr 构造一个带 code/reason/message 的业务错误
func NewErr(code int, reason, msg string) *errors.Error {
	return errors.New(code, reason, msg)
}

// error code
const (
	// 通用错误
	ErrCodeInvalidParams  = 400
	ErrCodeInternalServer = 500

	// 注册登录相关
	ErrCodeInvalidPhone           = 42201
	ErrCodePhoneNotFound          = 42202
	ErrCodePhoneAlreadyRegistered = 42203
	ErrCodeInvalidPassword        = 40101
	ErrCodeCreateTokenFailed      = 50001

	// 数据库、redis缓存相关
	ErrCodeDBQueryFailed    = 50002
	ErrCodeCreateUserFailed = 50003
)

// error reason
const (
	// 通用错误
	INVALID_PARAMS  = "INVALID_PARAMS"
	INTERNAL_SERVER = "INTERNAL_SERVER"

	// 注册登录相关
	INVALID_PHONE            = "INVALID_PHONE"
	PHONE_NOT_FOUND          = "PHONE_NOT_FOUND"
	PHONE_ALREADY_REGISTERED = "PHONE_ALREADY_REGISTERED"
	INVALID_PASSWORD         = "INVALID_PASSWORD"
	CREATE_TOKEN_FAILED      = "CREATE_TOKEN_FAILED"

	// 数据库相关
	DB_QUERY_FAILED    = "DB_QUERY_FAILED"
	CREATE_USER_FAILED = "CREATE_USER_FAILED"
)
