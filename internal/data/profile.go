package data

import (
	"context"
	"encoding/json"
	"errors"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/model"
	"strconv"
	"time"

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

	// Step 2: MySQL 事务
	err = r.data.DB().Transaction(func(tx *gorm.DB) error {
		// 插入关注关系
		follow := bizProfile.FollowTB{FollowerID: followerID, FolloweeID: followeeID}
		if err := tx.Create(&follow).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.New("already followed")
			}
			return err
		}

		// 更新双方的统计数据
		if err := tx.Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followerID).
			Update("follow_count", gorm.Expr("follow_count + 1")).Error; err != nil {
			return err
		}
		if err := tx.Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followeeID).
			Update("fan_count", gorm.Expr("fan_count + 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Step 3: Redis 写入
	if err := r.UpdateFollowCache(ctx, followerID, followeeID); err != nil {
		r.log.Errorf("Redis update failed, will repair later. follower=%d followee=%d err=%v",
			followerID, followeeID, err)
		// 将失败记录保存到修复队列
		_ = r.recordRepairTask(ctx, "follow", followerID, followeeID)
	}

	return nil
}

func (r *ProfileRepo) UpdateFollowCache(ctx context.Context, followerID uint32, followeeID uint32) error {
	keyFollowing := UserRedisKey(UserCachePrefix, "following", followerID)
	keyFollowers := UserRedisKey(UserCachePrefix, "followers", followeeID)

	return r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SAdd(ctx, keyFollowing, strconv.Itoa(int(followeeID)))
		pipe.SAdd(ctx, keyFollowers, strconv.Itoa(int(followerID)))
		pipe.Expire(ctx, keyFollowing, 24*time.Hour)
		pipe.Expire(ctx, keyFollowers, 24*time.Hour)
		return nil
	})
}

func (r *ProfileRepo) UnFollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// Step 1: 检查目标用户是否存在
	user, err := r.GetProfileByUserID(ctx, followeeID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("followee user not found")
	}

	// Step 2: MySQL 事务
	err = r.data.DB().Transaction(func(tx *gorm.DB) error {
		// 删除关注关系
		if err := tx.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
			Delete(&bizProfile.FollowTB{}).Error; err != nil {
			return err
		}

		// 更新计数
		if err := tx.Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followerID).
			Update("follow_count", gorm.Expr("follow_count - 1")).Error; err != nil {
			return err
		}
		if err := tx.Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followeeID).
			Update("fan_count", gorm.Expr("fan_count - 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Step 3: Redis 更新
	if err := r.updateUnfollowCache(ctx, followerID, followeeID); err != nil {
		r.log.Errorf("Redis update failed, will repair later. follower=%d followee=%d err=%v",
			followerID, followeeID, err)
		_ = r.recordRepairTask(ctx, "unfollow", followerID, followeeID)
	}

	return nil
}

func (r *ProfileRepo) UpdateUnfollowCache(ctx context.Context, followerID uint32, followeeID uint32) error {
	keyFollowing := UserRedisKey(UserCachePrefix, "following", followerID)
	keyFollowers := UserRedisKey(UserCachePrefix, "followers", followeeID)

	return r.data.Cache().Pipeline(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, keyFollowing, strconv.Itoa(int(followeeID)))
		pipe.SRem(ctx, keyFollowers, strconv.Itoa(int(followerID)))
		return nil
	})
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

// 定时修复
type RepairTask struct {
	Action     string `json:"action"`      // follow 或 unfollow
	FollowerID uint32 `json:"follower_id"` // 发起者
	FolloweeID uint32 `json:"followee_id"` // 被关注者
}

// cache 相关的需要修正
func (r *ProfileRepo) RecordRepairTask(ctx context.Context, action string, followerID, followeeID uint32) error {
	task := RepairTask{
		Action:     action,
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	data, _ := json.Marshal(task)
	return r.data.RDB().LPush(ctx, "follow:repair:queue", data).Err()
}

func (r *ProfileRepo) RepairFollowCache(ctx context.Context) {
	for {
		result, err := r.data.RDB().RPop(ctx, "follow:repair:queue").Result()
		if err == redis.Nil {
			// 没有任务
			return
		} else if err != nil {
			r.log.Errorf("Repair queue pop error: %v", err)
			return
		}

		var task RepairTask
		if err := json.Unmarshal([]byte(result), &task); err != nil {
			r.log.Errorf("Invalid repair task: %v", err)
			continue
		}

		// 修复 Redis
		if task.Action == "follow" {
			_ = r.updateFollowCache(ctx, task.FollowerID, task.FolloweeID)
		} else if task.Action == "unfollow" {
			_ = r.updateUnfollowCache(ctx, task.FollowerID, task.FolloweeID)
		}
	}
}
