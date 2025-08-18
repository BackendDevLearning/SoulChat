package data

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"
	"kratos-realworld/internal/service"

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

func (r *UserRepo) CreateUser(ctx context.Context, userRegister *bizUser.UserTB) error {
	rv := r.data.DB().Create(userRegister)
	if rv.Error != nil {
		return fmt.Errorf("failed to create user: %w", rv.Error)
	}
	return nil
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	redisKey := service.UserInfoPrefix + phone
	val, ok, err := r.data.Cache().Get(context.Background(), redisKey)

	if err != nil {
		return nil, err
	}

	fmt.Println(ok, val)

	u := &bizUser.UserTB{}
	result := r.data.DB().Where("phone = ?", phone).First(u)

	// 没查到用户，不算错误，返回nil, nil
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	// 数据库报错，如断开连接
	if result.Error != nil {
		return nil, result.Error
	}

	// 成功查询到用户，返回查询结果
	return u, nil
}
