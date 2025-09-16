package biz

import (
	"context"
	"errors"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/model"
	"kratos-realworld/internal/pkg/middleware/auth"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
)

type ProfileUsecase struct {
	pr bizProfile.ProfileRepo
	tx model.Transaction

	jwtc *conf.JWT
	log  *log.Helper
}

func NewProfileUsecase(pr bizProfile.ProfileRepo, tx model.Transaction, jwtc *conf.JWT, logger log.Logger) *ProfileUsecase {
	return &ProfileUsecase{
		pr:   pr,
		tx:   tx,
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
		return nil, errors.New("string convert error")
	}

	if uint32(userID) == uint32(tID) {
		return nil, NewErr(ErrCodeFollowFailed, FOLLOW_USER_FAILED, "user cannot follow themselves")
	}

	// 先判断是否已经关注了
	isFollow, _ := pc.pr.CheckFollow(ctx, uint32(userID), uint32(tID))
	if isFollow {
		return nil, NewErr(ErrCodeFollowFailed, FOLLOW_USER_FAILED, "follow relationship already existed")
	}

	var followCount, fanCount uint32

	err = pc.tx.InTx(ctx, func(ctx context.Context) error {
		if err := pc.pr.FollowUser(ctx, uint32(userID), uint32(tID)); err != nil {
			return NewErr(ErrCodeFollowFailed, FOLLOW_USER_FAILED, "failed to insert follow relationship")
		}

		var err error
		if followCount, err = pc.pr.IncrementFollowCount(ctx, uint32(userID), 1); err != nil {
			return NewErr(ErrCodeFollowFailed, FOLLOW_USER_FAILED, "failed to increase follower follow counts")
		}

		if fanCount, err = pc.pr.IncrementFanCount(ctx, uint32(tID), 1); err != nil {
			return NewErr(ErrCodeFollowFailed, FOLLOW_USER_FAILED, "failed to increase followee fan counts")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UserFollowFanReply{
		SelfID:      uint32(userID),
		FollowCount: followCount,
		TargetID:    uint32(tID),
		FanCount:    fanCount,
	}, nil

}

func (pc *ProfileUsecase) UnfollowUser(ctx context.Context, targetID string) (*UserFollowFanReply, error) {
	userID := auth.FromContext(ctx).UserID
	// 参数：字符串, 进制(10), 位数(32)
	tID, err := strconv.ParseUint(targetID, 10, 32)
	if err != nil {
		return nil, errors.New("string convert error")
	}

	if uint32(userID) == uint32(tID) {
		return nil, NewErr(ErrCodeUnfollowFailed, UNFOLLOW_USER_FAILED, "user cannot unfollow themselves")
	}

	isFollow, _ := pc.pr.CheckFollow(ctx, uint32(userID), uint32(tID))
	if !isFollow {
		return nil, NewErr(ErrCodeUnfollowFailed, UNFOLLOW_USER_FAILED, "the follow relationship does not exist")
	}

	var followCount, fanCount uint32

	err = pc.tx.InTx(ctx, func(ctx context.Context) error {
		if err := pc.pr.UnfollowUser(ctx, uint32(userID), uint32(tID)); err != nil {
			return NewErr(ErrCodeUnfollowFailed, UNFOLLOW_USER_FAILED, "failed to delete follow relationship")
		}

		var err error
		if followCount, err = pc.pr.IncrementFollowCount(ctx, uint32(userID), -1); err != nil {
			return NewErr(ErrCodeUnfollowFailed, UNFOLLOW_USER_FAILED, "failed to decrease follower follow counts")
		}

		if fanCount, err = pc.pr.IncrementFanCount(ctx, uint32(tID), -1); err != nil {
			return NewErr(ErrCodeUnfollowFailed, UNFOLLOW_USER_FAILED, "failed to decrease followee fan counts")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &UserFollowFanReply{
		SelfID:      uint32(userID),
		FollowCount: followCount,
		TargetID:    uint32(tID),
		FanCount:    fanCount,
	}, nil

}

func (pc *ProfileUsecase) GetRelationship(ctx context.Context, targetID string) (*UserRelationshipReply, error) {
	userID := auth.FromContext(ctx).UserID
	// 参数：字符串, 进制(10), 位数(32)
	tID, err := strconv.ParseUint(targetID, 10, 32)
	if err != nil {
		return nil, errors.New("string convert error")
	}

	var isFollowing, isFollowedBy, isMutual, isBlocked, isBlockedBy, isFriend bool

	// 判断是否关注 userID -> targetID
	isFollowing, err = pc.pr.CheckFollow(ctx, uint32(userID), uint32(tID))

	// 判断是否被关注 targetID -> userID
	isFollowedBy, err = pc.pr.CheckFollow(ctx, uint32(tID), uint32(userID))

	// 是否相互关注 userID <-> targetID
	isMutual = isFollowing && isFollowedBy

	// 是否拉黑 userID -> targetID
	isBlocked, err = pc.pr.CheckBlock(ctx, uint32(userID), uint32(tID))

	// 是否被拉黑 targetID -> userID
	isBlockedBy, err = pc.pr.CheckBlock(ctx, uint32(tID), uint32(userID))

	// 是否是好友 userID <-> targetID
	isFriend, err = pc.pr.CheckFriend(ctx, uint32(userID), uint32(tID))

	return &UserRelationshipReply{
		IsFollowing:  isFollowing,
		IsFollowedBy: isFollowedBy,
		IsMutual:     isMutual,
		IsBlocked:    isBlocked,
		IsBlockedBy:  isBlockedBy,
		IsFriend:     isFriend,
	}, nil
}

func (pc *ProfileUsecase) CanAddFriend(ctx context.Context, targetID string) (bool, error) {
	//userID := auth.FromContext(ctx).UserID
	//// 参数：字符串, 进制(10), 位数(32)
	//tID, err := strconv.ParseUint(targetID, 10, 32)
	//if err != nil {
	//	return nil, errors.New("string convert error")
	//}
	//
	//res, err := pc.pr.CanAddFriend(ctx, userID, tID)
	//if err != nil {
	//	return false, err
	//}
	//
	//return res, nil

	return true, nil
}
