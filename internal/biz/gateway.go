package biz

import (
	"context"
	"fmt"
	bizProfile "kratos-realworld/internal/biz/profile"
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
	pr bizProfile.ProfileRepo

	jwtc *conf.JWT
	log  *log.Helper
}

func NewGateWayUsecase(ur bizUser.UserRepo, pr bizProfile.ProfileRepo, jwtc *conf.JWT, logger log.Logger) *GateWayUsecase {
	return &GateWayUsecase{
		ur:   ur,
		pr:   pr,
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

	// 首次注册创建用户后，同时创建当前用户的空白主页profile，后续可以更新完善信息
	userProfile := &bizProfile.ProfileTB{
		UserID: user.ID,
	}
	if err := gc.pr.CreateProfile(ctx, userProfile); err != nil {
		return nil, NewErr(ErrCodeCreateUserFailed, CREATE_USER_FAILED, "failed to create user profile")
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

func (gc *GateWayUsecase) UpdatePassword(ctx context.Context, phone, oldPassword, newPassword string) error {
	if !IsValidPhone(phone) {
		return NewErr(ErrCodeInvalidPhone, INVALID_PHONE, "invalid phone number format")
	}

	if oldPassword == newPassword {
		return NewErr(ErrCodeInvalidPassword, INVALID_PASSWORD, "new password cannot be the same as the old password")
	}

	dataPassword, err := gc.ur.GetPasswordByPhone(ctx, phone)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query user by phone")
	}
	if dataPassword == "" {
		return NewErr(ErrCodePhoneNotFound, PHONE_NOT_FOUND, "phone number not registered")
	}

	// 密码输入错误
	if !verifyPassword(dataPassword, oldPassword) {
		return NewErr(ErrCodeInvalidPassword, INVALID_PASSWORD, "password is incorrect")
	}

	newPasswordHash := hashPassword(newPassword)
	err = gc.ur.UpdatePassword(ctx, phone, newPasswordHash)
	if err != nil {
		return NewErr(ErrCodeUpdatePasswordFailed, UPDATE_PASSWORD_FAILED, "failed to update password")
	}

	return nil
}

func (gc *GateWayUsecase) generateToken(userID uint32) (string, error) {
	expire, err := time.ParseDuration(gc.jwtc.Expire)
	if err != nil {
		return "", fmt.Errorf("invalid JWT expire configuration: %w", err)
	}

	return auth.GenerateToken(gc.jwtc.Secret, userID, expire)
}
