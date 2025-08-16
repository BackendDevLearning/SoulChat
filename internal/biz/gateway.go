package biz

import (
	"errors"
	"fmt"
	"kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type GateWay struct {
	UserRepo user.UserRepo
	jwtc     *conf.JWT
	log      *log.Helper
}

func NewGatWayCase(ur user.UserRepo, jwtc *conf.JWT, logger log.Logger) *GateWay {
	return &GateWay{
		UserRepo: ur,
		jwtc:     jwtc,
		log:      logger,
	}
}

func (ur *GateWay) Register(ctx context.Context, username, phone, password string) error {
	userRegister := &user.UserRegisterTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: user.HashPassword(password),
	}
	if err := ur.CreateUser(ctx, userRegister); err != nil {
		fmt.Errorf("")
		return err
	}
	return nil
}

func (ur *GateWay) generateToken(userID uint) string {
	return auth.GenerateToken(ur.jwtc.Secret, userID)
}

func (r *GateWay) GetUserByPhone(ctx context.Context, phone string) error {
	u := new(bizUser.UserRegisterTB)
	result := r.data.DB().Where("phone = ?", phone).First(u)

	if result.Error != nil {
		return err
	}
	return nil
}

func (ul *GateWay) Login(ctx context.Context, phone, password string) error {
	if len(phone) == 0 {
		return nil, errors.New(422, "email", "cannot be empty")
	}
	u, err := ul.ul.GetUserByPhone(ctx, phone)
	if err != nil {
		return err
	}
	return nil
}
