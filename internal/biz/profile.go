package biz

import (
	"context"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/conf"
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
	return &UserFollowFanReply{}, nil
}

func (pc *ProfileUsecase) UnfollowUser(ctx context.Context, targetID string) (*UserFollowFanReply, error) {
	return &UserFollowFanReply{}, nil
}
