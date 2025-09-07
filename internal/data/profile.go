package data

import (
	"context"
	"errors"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/model"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type ProfileRepo struct {
	data *model.Data
	log  *log.Helper
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

func (r *ProfileRepo) FollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 防止关注的用户不存在
	user, err := r.GetProfileByUserID(ctx, followerID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
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

	// 更新数据库
	r.updateFollowCache(ctx, followerID, followeeID)

	return nil
}

func (r *ProfileRepo) UpdateFollowCache(ctx context.Context, followerID uint32, followeeID uint32) {
	redisKey_foller := UserRedisKey(UserCachePrefix, "following", followerID)
	redisKey_follee := UserRedisKey(UserCachePrefix, "followee", followeeID)

	err := r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(ctx, redisKey_foller, strconv.Itoa(int(followeeID)))
		pipe.SAdd(ctx, redisKey_follee, strconv.Itoa(int(followerID)))
		return nil
	}) // 获取管道对象

	if err != nil {
		r.logger.Errorf("redis pipeline failed, follower=%d, followee=%d, err=%v",
			followerID, followeeID, err)
	}
}

func (r *ProfileRepo) UnFollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 防止取关的用户不存在
	user, err := r.GetProfileByUserID(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 取关
	if err := r.data.DB().Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&bizProfile.FollowTB{}).Error; err != nil {
		return err
	}

	// 2. 更新计数（MySQL）
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followerID).
		Update("follow_count", gorm.Expr("follow_count - 1"))
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followeeID).
		Update("fan_count", gorm.Expr("fan_count - 1"))
	r.updateUnfollowCache(ctx, followerID, followeeID)
	return nil
}

func (r *ProfileRepo) UpdateUnfollowCache(ctx context.Context, followerID uint32, followeeID uint32) {
	redisKey_foller := UserRedisKey(UserCachePrefix, "following", followerID)
	redisKey_follee := UserRedisKey(UserCachePrefix, "followee", followeeID)

	err := r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, redisKey_foller, strconv.Itoa(int(followeeID)))
		pipe.SRem(ctx, redisKey_follee, strconv.Itoa(int(followerID)))
		return nil
	}) // 获取管道对象

	if err != nil {
		r.logger.Errorf("redis pipeline failed, follower=%d, followee=%d, err=%v",
			followerID, followeeID, err)
	}
}

func (r *ProfileRepo) CanAddFriendCache(ctx context.Context, userID uint32, followerID uint32) (bool, error) {
	redis := r.data.Cache()
	keyA := UserRedisKey(UserCachePrefix, "following", userID)
	keyB := UserRedisKey(UserCachePrefix, "following", followerID)

	isAFollowsB, err := redis.SIsMember(ctx, keyA, strconv.Itoa(int(userID))).Result()
	if err != nil {
		return false, err
	}

	isBFollowsA, err := redis.SIsMember(ctx, keyB, strconv.Itoa(int(followerID))).Result()
	if err != nil {
		return false, err
	}

	return isAFollowsB && isBFollowsA, nil
}

func (r *ProfileRepo) CanAddFriendSql(ctx context.Context, userID uint32, followerID uint32) (bool, error) {
	var cnt int64
	r.data.DB().Model(&bizProfile.FollowTB{}).
		Where("follower_id = ? AND followee_id = ?", followerID, userID).
		Count(&cnt)
	if cnt == 0 {
		return false, nil
	}
	return true, nil
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

func (r *ProfileRepo) IncrementFollowCount(ctx context.Context, userID uint, delta int) error {
	return nil
}

func (r *ProfileRepo) IncrementFanCount(ctx context.Context, userID uint, delta int) error {
	return nil
}
