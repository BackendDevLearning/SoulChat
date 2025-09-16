package data

import (
	"context"
	"errors"
	"fmt"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
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

func (r *ProfileRepo) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(model.TxKey).(*gorm.DB); ok {
		fmt.Println("使用事务")
		return tx
	}
	fmt.Println("使用全局DB")
	return r.data.DB()
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	rv := r.data.DB().Create(profile)
	if rv.Error != nil {
		return rv.Error
	}

	redisKey := UserRedisKey(UserCachePrefix, "Profile", profile.UserID)
	_ = HSetMultiple(ctx, r.data, r.log, redisKey, profile)

	return nil
}

func (r *ProfileRepo) GetProfileByUserID(ctx context.Context, userID uint32) (*bizProfile.ProfileTB, error) {
	profile := &bizProfile.ProfileTB{}
	redisKey := UserRedisKey(UserCachePrefix, "Profile", userID)
	err := HGetMultiple(ctx, r.data, r.log, redisKey, profile)
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

		_ = HSetMultiple(ctx, r.data, r.log, redisKey, profile)
	}
	return profile, nil
}

func (r *ProfileRepo) UpdateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	return nil
}

func (r *ProfileRepo) FollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 检查目标用户是否存在
	user, err := r.GetProfileByUserID(ctx, followerID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("followee user not found")
	}

	// 插入关注关系记录
	follow := bizProfile.FollowFanTB{FollowerID: followerID, FolloweeID: followeeID}
	if err := r.getDB(ctx).Create(&follow).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("follow relationship already existed")
		}
		return err
	}

	// Redis 写入
	if err := r.addFollowFanCache(ctx, followerID, followeeID); err != nil {
		r.log.Errorf("Follow and Fan relationship Redis add failed. follower=%d followee=%d err=%v", followerID, followeeID, err)
	}

	return nil
}

func (r *ProfileRepo) addFollowFanCache(ctx context.Context, followerID uint32, followeeID uint32) error {
	keyFollowList := UserRedisKey(UserCachePrefix, "FollowList", followerID) // 某人的关注列表
	keyFanList := UserRedisKey(UserCachePrefix, "FanList", followeeID)       // 某人的粉丝列表

	return r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(ctx, keyFollowList, fmt.Sprintf("%d", followeeID))
		pipe.SAdd(ctx, keyFanList, fmt.Sprintf("%d", followerID))
		pipe.Expire(ctx, keyFollowList, UserCacheTTL)
		pipe.Expire(ctx, keyFanList, UserCacheTTL)
		return nil
	})
}

func (r *ProfileRepo) UnfollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 检查目标用户是否存在
	user, err := r.GetProfileByUserID(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("followee user not found")
	}

	// 删除关注关系记录
	res := r.getDB(ctx).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&bizProfile.FollowFanTB{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("follow relationship not found, not followed before")
	}

	// Redis 更新
	if err := r.deleteFollowFanCache(ctx, followerID, followeeID); err != nil {
		r.log.Errorf("Follow and Fan relationship Redis delete failed. follower=%d followee=%d err=%v", followerID, followeeID, err)
	}

	return nil
}

func (r *ProfileRepo) deleteFollowFanCache(ctx context.Context, followerID uint32, followeeID uint32) error {
	keyFollowList := UserRedisKey(UserCachePrefix, "FollowList", followerID) // 某人的关注列表
	keyFanList := UserRedisKey(UserCachePrefix, "FanList", followeeID)       // 某人的粉丝列表

	return r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, keyFollowList, fmt.Sprintf("%d", followeeID))
		pipe.SRem(ctx, keyFanList, fmt.Sprintf("%d", followerID))
		return nil
	})
}

func (r *ProfileRepo) CanAddFriendCache(ctx context.Context, userID uint32, followerID uint32) (bool, error) {
	//redis := r.data.Cache()
	//keyA := UserRedisKey(UserCachePrefix, "following", userID)
	//keyB := UserRedisKey(UserCachePrefix, "following", followerID)
	//
	//isAFollowsB, err := redis.SIsMember(ctx, keyA, strconv.Itoa(int(userID))).Result()
	//if err != nil {
	//	return false, err
	//}
	//
	//isBFollowsA, err := redis.SIsMember(ctx, keyB, strconv.Itoa(int(followerID))).Result()
	//if err != nil {
	//	return false, err
	//}
	//
	//return isAFollowsB && isBFollowsA, nil

	return true, nil
}

func (r *ProfileRepo) CanAddFriendSql(ctx context.Context, userID uint32, followerID uint32) (bool, error) {
	var cnt int64
	r.data.DB().Model(&bizProfile.FollowFanTB{}).
		Where("follower_id = ? AND followee_id = ?", followerID, userID).
		Count(&cnt)
	if cnt == 0 {
		return false, nil
	}
	return true, nil
}

func (r *ProfileRepo) CheckFollow(ctx context.Context, userID uint32, targetID uint32) (bool, error) {
	keyFollowList := UserRedisKey(UserCachePrefix, "FollowList", userID) // 某人的关注列表
	isFollow, err := r.data.Cache().SIsMember(ctx, keyFollowList, fmt.Sprintf("%d", targetID))
	if err != nil {
		r.log.Errorf("Redis SIsMember error: %v, fallback to DB", err)
	} else if isFollow {
		r.data.Cache().Expire(ctx, keyFollowList, UserCacheTTL)
		r.log.Debugf("get data from cache successfully, refreshed TTL to %s for key %s", UserCacheTTL, keyFollowList)
		return true, nil
	} else {
		r.log.Debugf("Key %s not found in Redis, fallback to DB", keyFollowList)
	}

	// 缓存中没有再去查mysql里面的关注粉丝关系表
	var cnt int64
	err = r.data.DB().Model(&bizProfile.FollowFanTB{}).
		Where("follower_id = ? AND followee_id = ?", userID, targetID).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (r *ProfileRepo) CheckBlock(ctx context.Context, userID uint32, targetID uint32) (bool, error) {
	return false, nil
}

func (r *ProfileRepo) CheckFriend(ctx context.Context, userID uint32, targetID uint32) (bool, error) {
	return false, nil
}

func (r *ProfileRepo) IncrementFollowCount(ctx context.Context, userID uint32, delta int) (uint32, error) {
	var newFollowCount uint32
	err := r.getDB(ctx).Model(&bizProfile.ProfileTB{}).
		Where("user_id = ?", userID).
		UpdateColumn("follow_count", gorm.Expr("follow_count + ?", delta)).
		Error
	if err != nil {
		return 0, err
	}

	err = r.getDB(ctx).Model(&bizProfile.ProfileTB{}).
		Where("user_id = ?", userID).
		Select("follow_count").
		Take(&newFollowCount).Error
	if err != nil {
		return 0, err
	}

	redisKey := UserRedisKey(UserCachePrefix, "Profile", userID)
	// 先查询这个用户的profile是否已经在redis缓存中
	_, err = r.data.Cache().HGet(ctx, redisKey, "follow_count")
	// 用户的profile不在redis里面：上面已经直接更新了mysql的fan_count字段，感觉没必要把全部的profile hash存到redis里面
	if err != nil {
		r.log.Warnf("failed to get oldFollowCount from cache, %v", err)
		return newFollowCount, nil
	}
	// 用户的profile已经被缓存到redis里面：直接修改里面的follow_count字段，更新为最新值
	if _, err := r.data.Cache().HSet(ctx, redisKey, "follow_count", newFollowCount); err != nil {
		r.log.Warnf("failed to update follow_count to cache, %v", err)
	} else {
		r.data.Cache().Expire(ctx, redisKey, UserCacheTTL)
		r.log.Debugf("Profile follow_count cached update successfully, set TTL to %s for key %s", UserCacheTTL, redisKey)
	}

	return newFollowCount, nil
}

func (r *ProfileRepo) IncrementFanCount(ctx context.Context, userID uint32, delta int) (uint32, error) {
	var newFanCount uint32
	err := r.getDB(ctx).Model(&bizProfile.ProfileTB{}).
		Where("user_id = ?", userID).
		UpdateColumn("fan_count", gorm.Expr("fan_count + ?", delta)).Error
	if err != nil {
		return 0, err
	}

	err = r.getDB(ctx).Model(&bizProfile.ProfileTB{}).
		Where("user_id = ?", userID).
		Select("fan_count").
		Take(&newFanCount).Error
	if err != nil {
		return 0, err
	}

	redisKey := UserRedisKey(UserCachePrefix, "Profile", userID)
	_, err = r.data.Cache().HGet(ctx, redisKey, "fan_count")
	if err != nil {
		r.log.Warnf("failed to get oldFollowCount from cache, %v", err)
		return newFanCount, nil
	}
	if _, err := r.data.Cache().HSet(ctx, redisKey, "fan_count", newFanCount); err != nil {
		r.log.Warnf("failed to update fan_count to cache, %v", err)
	} else {
		r.data.Cache().Expire(ctx, redisKey, UserCacheTTL)
		r.log.Debugf("Profile fan_count cached update successfully, set TTL to %s for key %s", UserCacheTTL, redisKey)
	}

	return newFanCount, nil
}
