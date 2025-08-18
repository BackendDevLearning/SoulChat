package biz

import (
	"context"
	"fmt"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
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

type RegisterReply struct {
	Phone    string
	UserName string
	Token    string
}

func (gc *GateWayUsecase) Register(ctx context.Context, username string, phone string, password string) (*RegisterReply, error) {
	// 构建新用户
	userRegister := &bizUser.UserTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: hashPassword(password),
	}

	// 先查数据库：手机号是否已存在
	existing, err := gc.ur.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by phone: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("phone already registered")
	}

	// 插入用户
	if err := gc.ur.CreateUser(ctx, userRegister); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 插入成功，返回数据库里刚创建的用户信息和Token
	return &RegisterReply{
		Phone:    phone,
		UserName: username,
		Token:    gc.generateToken(userRegister.ID),
	}, nil
}

func (gc *GateWayUsecase) Login(ctx context.Context, phone string, password string) (string, error) {
	res := ""
	if len(phone) == 0 {
		res = "phone cannot be empty"
		return res, errors.New(422, "PHONE_EMPTY", "phone cannot be empty")
	}

	//user, err := cache.
	//if err != nil {
	//	return "", err
	//}
	return res, nil
}

func (gc *GateWayUsecase) generateToken(userID uint) string {
	return auth.GenerateToken(gc.jwtc.Secret, userID)
}
