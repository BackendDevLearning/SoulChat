package data

import (
	"context"
	"errors"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type ProfileRepo struct {
	data *model.Data
	log  *log.Helper
}

func (r *ProfileRepo) CheckFollowTogether(ctx context.Context, followerID uint32, followeeID uint32) (bool, error) {
	var cnt1 int64
	var cnt2 int64
	r.data.DB().Model(&FollowTB{}).
		Where("follower_id = ? AND followee_id = ?", followeeID, followerID).
		Count(&cnt1)

	r.data.DB().Model(&FollowTB{}).
		Where("follower_id = ? AND followee_id = ?", followeeID, followerID).
		Count(&cnt2)

	if cnt1 > 0 && cnt2 > 0 {
		return true, nil
	}
	return false, nil
}

func (r *ProfileRepo) FollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 1. 插入关注关系
	follow := bizProfile.FollowTB{FollowerID: followerID, FolloweeID: followeeID}
	if err := r.data.DB().Create(&follow).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("already followed")
		}
		return err
	}

	// 2. 更新双方 profile
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", userID).
		Update("follow_count", gorm.Expr("follow_count + 1"))
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", targetID).
		Update("fan_count", gorm.Expr("fan_count + 1"))

	return nil
}

func NewProfileRepo(data *model.Data, logger log.Logger) bizProfile.ProfileRepo {
	return &ProfileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	rv := r.data.DB().Create(profile)
	if rv.Error != nil {
		return rv.Error
	}

	redisKey := UserRedisKey(UserCachePrefix, "Profile", profile.UserID)
	_ = HSetStruct(ctx, r.data, r.log, redisKey, profile)

	return nil
}

func (r *ProfileRepo) GetProfileByUserID(ctx context.Context, userID uint32) (*bizProfile.ProfileTB, error) {
	profile := &bizProfile.ProfileTB{}
	redisKey := UserRedisKey(UserCachePrefix, "Profile", userID)
	err := HGetStruct(ctx, r.data, r.log, redisKey, profile)
	if err != nil {
		r.log.Warnf("failed to get from cache, fallback to DB: %v", err)
	}
	// 缓存没有命中
	if profile.SysCreated == nil {
		result := r.data.DB().Where("user_id = ?", userID).First(profile)

		// 没查到用户的profile，不算错误，返回nil, gorm.ErrRecordNotFound
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		// 数据库报错，如断开连接
		if result.Error != nil {
			return nil, result.Error
		}

		_ = HSetStruct(ctx, r.data, r.log, redisKey, profile)
	}
	return profile, nil
}

func (r *ProfileRepo) UpdateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	return nil
}

func (r *ProfileRepo) IncrementFollowCount(ctx context.Context, userID uint, delta int) error {
	return nil
}

func (r *ProfileRepo) IncrementFanCount(ctx context.Context, userID uint, delta int) error {
	return nil
}
