package data

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"
	"kratos-realworld/internal/service"
)

type UserRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewUserRepo(data *model.Data, logger log.Logger) bizUser.UserRepo {
	return &UserRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	redisKey := service.UserInfoPrefix + phone
	val, ok, err := r.data.Cache().Get(context.Background(), redisKey)

	if err != nil {
		return nil, err
	}

	fmt.Println(ok, val)

	u := new(bizUser.UserTB)
	result := r.data.DB().Where("phone = ?", phone).First(u)

	if result.Error != nil {
		return u, result.Error
	}
	return u, nil
}

func (r *UserRepo) CreateUser(ctx context.Context, userRegister *bizUser.UserTB) (string, error) {
	rv := r.data.DB().Create(userRegister)
	if rv.Error != nil {
		return "data createuser failed, in data", rv.Error
	}
	return "", nil
}
