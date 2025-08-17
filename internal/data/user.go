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

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	u := new(bizUser.UserTB)
	result := r.data.DB().Where("phone = ?", phone).First(u)

	if result.Error != nil {
		return "phone is nil", result.Error
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
