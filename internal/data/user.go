package data

import (
	"context"
	"encoding/json"
	"errors"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
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
		return rv.Error
	}

	redisKey := UserRedisKey(UserCachePrefix, userRegister.ID)
	redisBytes, err := json.Marshal(userRegister)
	err = r.data.Cache().Set(ctx, redisKey, string(redisBytes), UserCacheTTL)
	if err != nil {
		r.log.Warnf("failed to write data to cache, but DB query succeeded: %v", err)
	} else {
		r.log.Debugf("data cached successfully")
	}

	return nil
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	redisKey := UserRedisKey(UserCachePrefix, phone)
	val, ok, err := r.data.Cache().Get(ctx, redisKey)
	if err != nil {
		r.log.Warnf("failed to get from cache, fallback to DB: %v", err)
	}

	if ok {
		var user bizUser.UserTB
		if err := json.Unmarshal([]byte(val), &user); err != nil {
			r.log.Warnf("failed to unmarshal cached user, fallback to DB: %v", err)
		}
		r.log.Debugf("get data from cache successfully")
		return &user, nil
	}

	user := &bizUser.UserTB{}
	result := r.data.DB().Where("phone = ?", phone).First(user)

	// 没查到用户，不算错误，返回nil, gorm.ErrRecordNotFound
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}

	// 数据库报错，如断开连接
	if result.Error != nil {
		return nil, result.Error
	}

	redisBytes, err := json.Marshal(user)
	err = r.data.Cache().Set(ctx, redisKey, string(redisBytes), UserCacheTTL)
	if err != nil {
		r.log.Warnf("failed to write data to cache, but DB query succeeded: %v", err)
	} else {
		r.log.Debugf("data cached successfully")
	}

	return user, nil
}
