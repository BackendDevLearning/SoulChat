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

func (r *UserRepo) GetPasswordByPhone(ctx context.Context, phone string) (string, error) {
	redisKey := UserRedisKey(UserCachePrefix, phone)
	val, ok, err := r.data.Cache().Get(ctx, redisKey)
	if err != nil {
		r.log.Warnf("failed to get password from cache, fallback to DB: %v", err)
	}

	if ok {
		// log 打印一下 success 和 值
		return val, nil
	}

	// 2. 缓存没有 → 查数据库
	//result 是一个 *gorm.DB，里面有：
	//result.Error → 是否有错误（连接失败 / SQL 错误）
	//result.RowsAffected → 影响的行数（0 表示没查到）
	var password string
	result := r.data.DB().Model(&bizUser.UserTB{}).Select("password").Where("phone = ?", phone).Scan(&password)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", gorm.ErrRecordNotFound
	}
	if result.Error != nil {
		return "", result.Error
	}

	// 3. 回写缓存
	if err := r.data.Cache().Set(ctx, redisKey, password, UserCacheTTL); err != nil {
		r.log.Warnf("failed to write password to cache, but DB query succeeded: %v", err)
	} else {
		r.log.Debugf("password cached successfully, phone=%s", phone)
	}

	return password, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, phone string, new_password string) (string, error) {

	result := r.data.DB().Model(&bizUser.UserTB{}).Where("phone = ?", phone).Update("password", new_password)

	if result.Error != nil {
		return "", result.Error
	}

	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}

	// 同步更新缓存
	redisKey := UserRedisKey(UserCachePrefix, phone)
	if err := r.data.Cache().Delete(ctx, redisKey); err != nil {
		r.log.Warnf("failed to delete password in cache: %v", err)
	} else {
		r.log.Debugf("password delete in cache successfully, phone=%s", phone)
	}

	return new_password, nil

}
