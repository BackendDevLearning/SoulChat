package user

import (
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"

	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type userRegisterRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewUserRegisterRepo(data *model.Data, logger log.Logger) bizUser.UserRegisterRepo {
	return &userRegisterRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userRegisterRepo) CreateUser(ctx context.Context, userRegister *bizUser.UserRegisterTB) error {
	rv := r.data.DB().Create(userRegister)
	return rv.Error
}

//func (r *UserRegisterRepo) CreateUser(ctx context.Context, u *user.UserRegister) error {
//	db := r.data.DB()
//	if err := db.Model(u).Create(u).Error; err != nil {
//		r.log.Errorf("CreateUser fail: %v", err)
//		return fmt.Errorf("CreateUser fail: %v", err)
//	}
//	r.log.Infof("insert success")
//	return nil
//}
//
//func (r *UserRegisterRepo) UpdateByCache(u *user.UserRegister) error {
//	// TODO: Implement cache update logic
//	r.log.Infof("UpdateByCache called for user: %s", u.UserName)
//	return nil
//}
//
//func (r *UserRegisterRepo) Load(u *user.UserRegister) error {
//	// TODO: Implement load logic
//	r.log.Infof("Load called for user: %s", u.UserName)
//	return nil
//}
