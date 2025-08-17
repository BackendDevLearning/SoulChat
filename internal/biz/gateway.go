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

func NewGatWayUsecase(ur bizUser.UserRepo, jwtc *conf.JWT, logger log.Logger) *GateWayUsecase {
	return &GateWayUsecase{
		ur:   ur,
		jwtc: jwtc,
		log:  log.NewHelper(logger),
	}
}

func (gc *GateWayUsecase) Register(ctx context.Context, username string, phone string, password string) error {
	userRegister := &bizUser.UserTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: hashPassword(password),
	}
	if err := gc.ur.CreateUser(ctx, userRegister); err != nil {
		fmt.Errorf("")
		return err
	}
	return nil
}

func (ur *GateWayUsecase) generateToken(userID uint) string {
	return auth.GenerateToken(ur.jwtc.Secret, userID)
}

func (gc *GateWayUsecase) Login(ctx context.Context, phone string, password string) error {
	if len(phone) == 0 {
		return errors.New(422, "PHONE_EMPTY", "phone cannot be empty")
	}
	u, err := gc.ur.GetUserByPhone(ctx, phone, password)
	if err != nil {
		return err
	}
	return nil
}

func (gc *GateWayUsecase) GetUserByPhone(ctx context.Context, phone string) error {
	u := new(bizUser.UserTB)
	result := gc.ur.GetUserByPhone(ctx, phone)

	if err := result.Error; err != nil {
		return err
	}
	return nil
}
