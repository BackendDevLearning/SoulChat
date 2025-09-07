package biz

import (
	"context"
	"errors"
	"fmt"
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
	userID := auth.FromContext(ctx).UserID
	// 参数：字符串, 进制(10), 位数(32)
	tID, err := strconv.ParseUint(targetID, 10, 32)
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return nil, errors.New("string convert error")
	}

	er := pc.pr.FollowUser(ctx, uint32(tID), uint32(userID))
	if er != nil {
		return nil, er
	}

	profile, err := pc.pr.GetProfileByUserID(ctx, uint32(userID))
	if err != nil {
		return nil, err
	}

	// 3. 检查是否形成双向关注
	//together, err := pc.pr.CheckFollowTogether(ctx, uint32(tID), uint32(userID))
	//if err != nil {
	//	return nil, err
	//}

	return &UserFollowFanReply{
		SelfID:      uint32(userID),
		FollowCount: profile.FollowCount,
		TargetID:    uint32(tID),
		FanCount:    profile.FanCount,
	}, nil

}

func (pc *ProfileUsecase) UnfollowUser(ctx context.Context, targetID string) (*UserFollowFanReply, error) {
	userID := auth.FromContext(ctx).UserID
	// 参数：字符串, 进制(10), 位数(32)
	tID, err := strconv.ParseUint(targetID, 10, 32)
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return nil, errors.New("string convert error")
	}

	er := pc.pr.UnFollowUser(ctx, uint32(tID), uint32(userID))
	if er != nil {
		return nil, er
	}

	profile, err := pc.pr.GetProfileByUserID(ctx, uint32(userID))
	if err != nil {
		return nil, err
	}

	return &UserFollowFanReply{
		SelfID:      uint32(userID),
		FollowCount: profile.FollowCount,
		TargetID:    uint32(tID),
		FanCount:    profile.FanCount,
	}, nil

}

func (pc *ProfileUsecase) CanAddFriend(ctx context.Context, user_1 int32, user_2 int32) (bool, error) {
	res, err := pc.pr.CanAddFriend(ctx, uint32(user_1), uint32(user_2))
	if err != nil {
		return false, err
	}

	return res, nil
}
