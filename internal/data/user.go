package data

import (
	"context"
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

	redisKey := UserRedisKey(UserCachePrefix, "ID", userRegister.ID)
	_ = HSetStruct(ctx, r.data, r.log, redisKey, userRegister)

	return nil
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	user := &bizUser.UserTB{}
	redisKey := UserRedisKey(UserCachePrefix, "Phone", phone)

	err := HGetStruct(ctx, r.data, r.log, redisKey, user)
	if err != nil {
		r.log.Warnf("failed to get from cache, fallback to DB: %v", err)
	} else {
		return user, nil
	}

	result := r.data.DB().Where(&bizUser.UserTB{Phone: phone}).First(user)

	// 没查到用户，不算错误，返回nil, gorm.ErrRecordNotFound
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}

	// 数据库报错，如断开连接
	if result.Error != nil {
		return nil, result.Error
	}

	// 即使缓存写入失败，也不会影响主流程，仅打印日志，不把错误传到service层
	_ = HSetStruct(ctx, r.data, r.log, redisKey, user)

	return user, nil
}

func (r *UserRepo) GetPasswordByPhone(ctx context.Context, phone string) (string, error) {
	redisKey := UserRedisKey(UserCachePrefix, "Phone", phone)
	password, err := r.data.Cache().HGet(ctx, redisKey, "PassWord")
	if err != nil {
		r.log.Warnf("failed to get password from cache, fallback to DB: %v", err)
	} else {
		return password, nil
	}

	// 2. 缓存没有 → 查数据库
	//result 是一个 *gorm.DB，里面有：
	//result.Error → 是否有错误（连接失败 / SQL 错误）
	//result.RowsAffected → 影响的行数（0 表示没查到）
	res := r.data.DB().Model(&bizUser.UserTB{}).Select("password").Where(&bizUser.UserTB{Phone: phone}).Scan(&password)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return "", gorm.ErrRecordNotFound
	}
	if res.Error != nil {
		return "", res.Error
	}

	// 3. 回写缓存
	if _, err := r.data.Cache().HSet(ctx, redisKey, "PassWord", password); err != nil {
		r.log.Warnf("failed to write password to cache, but DB query succeeded: %v", err)
	} else {
		r.data.Cache().Expire(ctx, redisKey, UserCacheTTL)
		r.log.Debugf("password cached successfully, set TTL to %s for key %s", UserCacheTTL, redisKey)
	}

	return password, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, phone string, newPasswordHash string) error {
	result := r.data.DB().Model(&bizUser.UserTB{}).Where(&bizUser.UserTB{Phone: phone}).Update("password", newPasswordHash)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 同步更新缓存
	redisKey := UserRedisKey(UserCachePrefix, "Phone", phone)
	if _, err := r.data.Cache().HDel(ctx, redisKey, "PassWord"); err != nil {
		r.log.Warnf("failed to delete password in cache: %v", err)
	} else {
		r.log.Debugf("password delete in cache successfully, phone=%s", phone)
	}

	return nil
}
