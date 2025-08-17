package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
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

func (gc *GateWayUsecase) Register(ctx context.Context, username string, phone string, password string) (string, error) {
	userRegister := &bizUser.UserTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: hashPassword(password),
	}

	name_db, err := gc.ur.GetUserByPhone(ctx, phone)

	fmt.Println("smc", name_db)

	if err != nil {
		return "gc.ur.GetUserByPhone(ctx, phone) error", err
	}

	if name_db != nil {
		return "Phont is used", nil
	}

	if result, err := gc.ur.CreateUser(ctx, userRegister); err != nil {
		fmt.Errorf("")
		return result, err
	}
	return "", nil
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

func (ur *GateWayUsecase) generateToken(userID uint) string {
	return auth.GenerateToken(ur.jwtc.Secret, userID)
}
