package user

import (
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"

	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type userLoginRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewUserLoginRepo(data *model.Data, logger log.Logger) bizUser.UserLoginRepo {
	return &userLoginRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *userLoginRepo) GetUserByPhone(ctx context.Context, phone string) (rv *bizUser.UserLoginReply, err error) {
	u := new(bizUser.UserRegisterTB)
	result := r.data.DB().Where("phone = ?", phone).First(u)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("user", "not found by email")
	}
	if result.Error != nil {
		return nil, err
	}
	return &bizUser.UserLoginReply{
		ID:           u.ID,
		Phone:        u.Phone,
		Username:     u.UserName,
		Bio:          u.Bio,
		Image:        u.Image,
		PasswordHash: u.PasswordHash,
	}, nil
}

//func (r *UserLoginRepo) CreateUser(ctx context.Context, u *user.UserLogin) error {
//	db := r.data.DB()
//	if err := db.Model(u).Create(u).Error; err != nil {
//		r.log.Errorf("CreateUser fail: %v", err)
//		return fmt.Errorf("CreateUser fail: %v", err)
//	}
//	r.log.Infof("insert success")
//	return nil
//}
//
//func (r *UserLoginRepo) UpdateByCache(u *user.UserLogin) error {
//	// TODO: Implement cache update logic
//	r.log.Infof("UpdateByCache called for user: %s", u.UserName)
//	return nil
//}
//
//func (r *UserLoginRepo) Load(u *user.UserLogin) error {
//	// TODO: Implement load logic
//	r.log.Infof("Load called for user: %s", u.UserName)
//	return nil
//}
