package data

import (
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"

	"context"
	"github.com/go-kratos/kratos/v2/log"
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

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) error {
	u := new(bizUser.UserRegisterTB)
	result := r.data.DB().Where("phone = ?", phone).First(u)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) CreateUser(ctx context.Context, userRegister *bizUser.UserRegisterTB) error {
	rv := r.data.DB().Create(userRegister)
	return rv.Error
}
