package biz

import (
	"context"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

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
		return nil, errors.New(422, "INVALID_PHONE", "invalid phone number format")
	}

	// 先查数据库：手机号是否已存在
	existing, err := gc.ur.GetUserByPhone(ctx, phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(500, "DB_QUERY_FAILED", "failed to query user by phone")
	}
	if existing != nil {
		return nil, errors.New(422, "PHONE_ALREADY_REGISTERED", "phone already registered")
	}

	// 插入用户
	if err := gc.ur.CreateUser(ctx, user); err != nil {
		return nil, errors.New(500, "CREATE_USER_FAILED", "failed to create user")
	}

	// 插入成功，返回数据库里刚创建的用户信息和Token
	return &UserRegisterReply{
		Phone:    phone,
		UserName: username,
		Token:    gc.generateToken(user.ID),
	}, nil
}

func (gc *GateWayUsecase) Login(ctx context.Context, phone string, password string) (*UserLoginReply, error) {
	user := &bizUser.UserTB{
		Phone:        phone,
		PasswordHash: hashPassword(password),
	}

	if !IsValidPhone(user.Phone) {
		return nil, errors.New(422, "INVALID_PHONE", "invalid phone number format")
	}

	res, err := gc.ur.GetUserByPhone(ctx, user.Phone)

	// 查询，判断用户是否已经注册
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New(500, "DB_QUERY_FAILED", "failed to query user by phone")
	}
	if res == nil {
		return nil, errors.New(422, "PHONE_NOT_FOUND", "phone number not registered")
	}

	// 密码输入错误
	if !verifyPassword(res.PasswordHash, password) {
		return nil, errors.New(401, "INVALID_PASSWORD", "password is incorrect")
	}

	return &UserLoginReply{
		Phone:    res.Phone,
		UserName: res.UserName,
		Token:    gc.generateToken(user.ID),
	}, nil
}

func (gc *GateWayUsecase) generateToken(userID uint) string {
	return auth.GenerateToken(gc.jwtc.Secret, userID)
}
