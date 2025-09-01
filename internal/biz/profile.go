package biz

import (
	"context"
	"gorm.io/gorm"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

type ProfileUsecase struct {
	pr bizProfile.ProfileRepo

	jwtc *conf.JWT
	log  *log.Helper
}

func NewProfileUsecase(pr bizProfile.ProfileRepo, jwtc *conf.JWT, logger log.Logger) *ProfileUsecase {
	return &ProfileUsecase{
		pr:   pr,
		jwtc: jwtc,
		log:  log.NewHelper(logger),
	}
}

func (pc *ProfileUsecase) GetProfile(ctx context.Context, userID string) (*UserProfileReply, error) {
	id, _ := strconv.ParseUint(userID, 10, 32)
	res, err := pc.pr.GetProfileByUserID(ctx, uint32(id))
	if err != nil {
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query profile by UserID")
	}

	return &UserProfileReply{
		UserID:            res.UserID,
		Tags:              res.Tags,
		FollowCount:       res.FollowCount,
		FanCount:          res.FanCount,
		ViewCount:         res.ViewCount,
		NoteCount:         res.NoteCount,
		ReceivedLikeCount: res.ReceivedLikeCount,
		CollectedCount:    res.CollectedCount,
		CommentCount:      res.CommentCount,
		LastLoginIP:       res.LastLoginIP,
		LastActive:        res.LastActive,
		Status:            res.Status,
	}, nil
}

func (pc *ProfileUsecase) FollowUser(ctx context.Context, targetID string) (*UserFollowFanReply, error) {
	userID := auth.FromContext(ctx)

	// 1. 插入关注关系
	follow := FollowTB{FollowerID: userID, FolloweeID: targetID}
	if err := pc.data.DB().Create(&follow).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("already followed")
		}
		return nil, err
	}

	// 2. 更新双方 profile
	pc.data.DB().Model(&ProfileTB{}).Where("user_id = ?", userID).
		Update("follow_count", gorm.Expr("follow_count + 1"))
	pc.data.DB().Model(&ProfileTB{}).Where("user_id = ?", targetID).
		Update("fan_count", gorm.Expr("fan_count + 1"))

	// 3. 检查是否形成双向关注
	var cnt int64
	pc.data.DB().Model(&FollowTB{}).
		Where("follower_id = ? AND followee_id = ?", targetID, userID).
		Count(&cnt)

	return &UserFollowFanReply{
		MutualFollow: cnt > 0,
	}, nil

	return &UserFollowFanReply{}, nil
}

func (pc *ProfileUsecase) UnfollowUser(ctx context.Context, targetID string) (*UserFollowFanReply, error) {
	userID := auth.FromContext(ctx)

	// 1. 删除关系
	pc.data.DB().Where("follower_id = ? AND followee_id = ?", userID, targetID).
		Delete(&FollowTB{})

	// 2. 更新双方 profile
	pc.data.DB().Model(&ProfileTB{}).Where("user_id = ?", userID).
		Update("follow_count", gorm.Expr("follow_count - 1"))
	pc.data.DB().Model(&ProfileTB{}).Where("user_id = ?", targetID).
		Update("fan_count", gorm.Expr("fan_count - 1"))

	// 3. 检查是否还存在互关
	var cnt int64
	pc.data.DB().Model(&FollowTB{}).
		Where("follower_id = ? AND followee_id = ?", targetID, userID).
		Count(&cnt)

	return &UserFollowFanReply{
		MutualFollow: cnt > 0,
	}, nil

	return &UserFollowFanReply{}, nil
}
