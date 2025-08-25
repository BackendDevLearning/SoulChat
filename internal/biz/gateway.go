package biz

import (
	"context"
	"fmt"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type GateWayUsecase struct {
	ur bizUser.UserRepo

	jwtc *conf.JWT
	log  *log.Helper
}

func NewGateWayUsecase(ur bizUser.UserRepo, jwtc *conf.JWT, logger log.Logger) *GateWayUsecase {
	return &GateWayUsecase{
		ur:   ur,
		jwtc: jwtc,
		log:  log.NewHelper(logger),
	}
}

func (gc *GateWayUsecase) Register(ctx context.Context, username string, phone string, password string) (*UserRegisterReply, error) {
	// 构建新用户
	user := &bizUser.UserTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: hashPassword(password),
	}

	// 验证输入phone是否有效
	if !IsValidPhone(user.Phone) {
		return nil, NewErr(ErrCodeInvalidPhone, INVALID_PHONE, "invalid phone number format")
	}

	// 先查数据库：手机号是否已存在
	existing, err := gc.ur.GetUserByPhone(ctx, phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query user by phone")
	}
	if existing != nil {
		return nil, NewErr(ErrCodePhoneAlreadyRegistered, PHONE_ALREADY_REGISTERED, "phone already registered")
	}

	// 插入用户
	if err := gc.ur.CreateUser(ctx, user); err != nil {
		return nil, NewErr(ErrCodeCreateUserFailed, CREATE_USER_FAILED, "failed to create user")
	}

	token, err := gc.generateToken(user.ID)
	if err != nil {
		return nil, NewErr(ErrCodeCreateTokenFailed, CREATE_TOKEN_FAILED, "failed to create token")
	}

	// 插入成功，返回数据库里刚创建的用户信息和Token
	return &UserRegisterReply{
		Phone:    phone,
		UserName: username,
		Token:    token,
	}, nil
}

func (gc *GateWayUsecase) Login(ctx context.Context, phone string, password string) (*UserLoginReply, error) {
	user := &bizUser.UserTB{
		Phone:        phone,
		PasswordHash: hashPassword(password),
	}

	if !IsValidPhone(user.Phone) {
		return nil, NewErr(ErrCodeInvalidPhone, INVALID_PHONE, "invalid phone number format")
	}

	res, err := gc.ur.GetUserByPhone(ctx, user.Phone)

	// 查询，判断用户是否已经注册
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query user by phone")
	}
	if res == nil {
		return nil, NewErr(ErrCodePhoneNotFound, PHONE_NOT_FOUND, "phone number not registered")
	}

	// 密码输入错误
	if !verifyPassword(res.PasswordHash, password) {
		return nil, NewErr(ErrCodeInvalidPassword, INVALID_PASSWORD, "password is incorrect")
	}

	token, err := gc.generateToken(user.ID)
	if err != nil {
		return nil, NewErr(ErrCodeCreateTokenFailed, CREATE_TOKEN_FAILED, "failed to create token")
	}

	return &UserLoginReply{
		Phone:    res.Phone,
		UserName: res.UserName,
		Token:    token,
	}, nil
}

func (gc *GateWayUsecase) UpdatePassword(ctx context.Context, new_password string, old_password string, phone string) (string, error) {
	if !IsValidPhone(phone) {
		return "invalid phone number format", nil
	}

	dataPassword, err := gc.ur.GetPasswordByPhone(ctx, phone)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "failed to query user by phone", nil
	}

	old_password = hashPassword(old_password)
	new_password = hashPassword(new_password)

	// 密码输入错误
	if !verifyPassword(dataPassword, old_password) {
		return "password is incorrect", nil
	}

	res, err := gc.ur.UpdatePassword(ctx, phone, new_password)

	if err != nil {
		return "update data error", err
	}

	return res, nil
}

func (gc *GateWayUsecase) generateToken(userID uint) (string, error) {
	expire, err := time.ParseDuration(gc.jwtc.Expire)
	if err != nil {
		return "", fmt.Errorf("invalid JWT expire configuration: %w", err)
	}

	return auth.GenerateToken(gc.jwtc.Secret, userID, expire)
}
