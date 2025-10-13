package data

import (
	"context"
	"errors"
	"fmt"
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

	redisKey := UserRedisKey(UserCachePrefix, "Phone", userRegister.Phone)
	_ = HSetMultiple(ctx, r.data, r.log, redisKey, userRegister)

	redisKeyID := UserRedisKey(UserCachePrefix, "ID", userRegister.ID)
	_ = HSetMultiple(ctx, r.data, r.log, redisKeyID, userRegister)

	return nil
}

func (r *UserRepo) GetUserByUserID(ctx context.Context, userID uint32) (*bizUser.UserTB, error) {
	user := &bizUser.UserTB{}
	redisKey := UserRedisKey(UserCachePrefix, "ID", userID)

	err := HGetMultiple(ctx, r.data, r.log, redisKey, user)
	if err != nil {
		r.log.Warnf("failed to get from cache, fallback to DB: %v", err)
	} else {
		return user, nil
	}

	result := r.data.DB().Where("id = ?", userID).First(user)

	// 没查到用户，不算错误，返回nil, gorm.ErrRecordNotFound
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}

	// 数据库报错，如断开连接
	if result.Error != nil {
		return nil, result.Error
	}

	// 即使缓存写入失败，也不会影响主流程，仅打印日志，不把错误传到service层
	_ = HSetMultiple(ctx, r.data, r.log, redisKey, user)

	return user, nil
}

func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*bizUser.UserTB, error) {
	user := &bizUser.UserTB{}
	redisKey := UserRedisKey(UserCachePrefix, "Phone", phone)

	err := HGetMultiple(ctx, r.data, r.log, redisKey, user)
	if err != nil {
		r.log.Warnf("failed to get from cache, fallback to DB: %v", err)
	} else {
		return user, nil
	}

	result := r.data.DB().Where("Phone = ?", phone).First(user)

	// 没查到用户，不算错误，返回nil, gorm.ErrRecordNotFound
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}

	// 数据库报错，如断开连接
	if result.Error != nil {
		return nil, result.Error
	}

	// 即使缓存写入失败，也不会影响主流程，仅打印日志，不把错误传到service层
	_ = HSetMultiple(ctx, r.data, r.log, redisKey, user)

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
	res := r.data.DB().Model(&bizUser.UserTB{}).Select("password").Where("Phone = ?", phone).Scan(&password)
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

func (r *UserRepo) UpdateUserPassword(ctx context.Context, phone string, newPasswordHash string) error {
	result := r.data.DB().Model(&bizUser.UserTB{}).Where("Phone = ?", phone).Update("password", newPasswordHash)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 同步更新缓存，两种更新策略：delete or update
	redisKey := UserRedisKey(UserCachePrefix, "Phone", phone)
	// 只删掉缓存hash里面的password字段，下次登录的时候，这个key对应的缓存hash还存在，但是取不到密码，会出密码无效的问题
	//if _, err := r.data.Cache().HDel(ctx, redisKey, "PassWord"); err != nil {
	if err := r.data.Cache().Delete(ctx, redisKey); err != nil {
		r.log.Warnf("failed to delete password in cache: %v", err)
	} else {
		r.log.Debugf("password delete in cache successfully, phone=%s", phone)
	}

	return nil
}

func (r *UserRepo) UpdateUserInfo(ctx context.Context, userID uint32, userInfo *bizUser.UpdateUserInfoFields) error {
	updateData := map[string]interface{}{}

	fmt.Println("userInfo", userInfo)

	if userInfo.Username != nil {
		updateData["UserName"] = *userInfo.Username
	}
	if userInfo.Gender != nil {
		updateData["Gender"] = *userInfo.Gender
	}
	if userInfo.Birthday != nil {
		updateData["Birthday"] = *userInfo.Birthday
	}
	if userInfo.Bio != nil {
		updateData["Bio"] = *userInfo.Bio
	}
	if userInfo.HeadImage != nil {
		updateData["HeadImage"] = *userInfo.HeadImage
	}
	if userInfo.CoverImage != nil {
		updateData["CoverImage"] = *userInfo.CoverImage
	}

	res := r.data.DB().Model(&bizUser.UserTB{}).Where("id = ?", userID).Updates(updateData)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 更新redis缓存
	redisKey := UserRedisKey(UserCachePrefix, "UserID", userID)
	// 先查询这个用户的profile是否已经在redis缓存中
	user := &bizUser.UserTB{}
	err := HGetMultiple(ctx, r.data, r.log, redisKey, user)
	// 用户信息不在redis里面
	if err != nil {
		// 查询更新后的完整信息
		if err := r.data.DB().First(&user, userID).Error; err != nil {
			return nil
		}
		// 将mysql更新后的数据写入redis缓存
		_ = HSetMultiple(ctx, r.data, r.log, redisKey, user)
		return nil
	}
	// 用户信息已经被缓存到redis里面：直接修改需要更新的字段
	_ = HSetMultiple(ctx, r.data, r.log, redisKey, updateData)

	return nil
}
