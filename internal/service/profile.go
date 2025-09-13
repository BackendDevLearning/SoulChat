package service

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/biz"

	"log"
)

func (cs *ConduitService) GetProfile(ctx context.Context, req *v1.GetProfileRequest) (*v1.GetProfileReply, error) {
	res, err := cs.pc.GetProfile(ctx, req.UserId)
	if err != nil {
		log.Printf("GetProfile err: %v", err)

		return &v1.GetProfileReply{
			Code: 1,
			Res:  ErrorToRes(err),
			Data: nil,
		}, nil
	}

	return &v1.GetProfileReply{
		Code: 0,
		Res:  ErrorToRes(err),
		Data: ConvertToProfileData(res),
	}, nil
}

func ConvertToProfileData(res *biz.UserProfileReply) *v1.ProfileData {
	var lastActiveProto *timestamppb.Timestamp
	if res.LastActive != nil {
		lastActiveProto = timestamppb.New(*res.LastActive)
	}
	return &v1.ProfileData{
		UserId:            res.UserID,
		Tags:              res.Tags,
		FollowCount:       res.FollowCount,
		FanCount:          res.FanCount,
		ViewCount:         res.ViewCount,
		NoteCount:         res.NoteCount,
		ReceivedLikeCount: res.ReceivedLikeCount,
		CollectedCount:    res.CollectedCount,
		CommentCount:      res.CommentCount,
		LastLoginIp:       res.LastLoginIP,
		LastActive:        lastActiveProto,
		Status:            res.Status,
	}
}

func (cs *ConduitService) FollowUser(ctx context.Context, req *v1.FollowUserRequest) (*v1.FollowFanReply, error) {
	res, err := cs.pc.FollowUser(ctx, req.TargetId)
	if err != nil {
		log.Printf("FollowUser err: %v", err)

		return &v1.FollowFanReply{
			Code: 1,
			Res:  ErrorToRes(err),
			Data: nil,
		}, nil
	}

	return &v1.FollowFanReply{
		Code: 0,
		Res:  ErrorToRes(err),
		Data: &v1.FollowFanData{
			SelfId:      res.SelfID,
			FollowCount: res.FollowCount,
			TargetId:    res.TargetID,
			FanCount:    res.FanCount,
		},
	}, nil
}

func (cs *ConduitService) UnfollowUser(ctx context.Context, req *v1.UnfollowUserRequest) (*v1.FollowFanReply, error) {
	res, err := cs.pc.UnfollowUser(ctx, req.TargetId)
	if err != nil {
		log.Printf("UnfollowUser err: %v", err)

		return &v1.FollowFanReply{
			Code: 1,
			Res:  ErrorToRes(err),
			Data: nil,
		}, nil
	}

	return &v1.FollowFanReply{
		Code: 0,
		Res:  ErrorToRes(err),
		Data: &v1.FollowFanData{
			SelfId:      res.SelfID,
			FollowCount: res.FollowCount,
			TargetId:    res.TargetID,
			FanCount:    res.FanCount,
		},
	}, nil
}

func (cs *ConduitService) GetRelationship(ctx context.Context, req *v1.RelationshipRequest) (*v1.RelationshipReply, error) {
	res, err := cs.pc.GetRelationship(ctx, req.TargetId)

	if err != nil {
		log.Printf("GetRelationship err: %v", err)

		return &v1.RelationshipReply{
			Code: 1,
			Res:  ErrorToRes(err),
			Data: nil,
		}, nil
	}

	return &v1.RelationshipReply{
		Code: 0,
		Res:  ErrorToRes(err),
		Data: &v1.RelationshipData{
			IsFollowing:  res.IsFollowing,
			IsFollowedBy: res.IsFollowedBy,
			IsMutual:     res.IsMutual,
			IsBlocked:    res.IsBlocked,
			IsBlockedBy:  res.IsBlockedBy,
			IsFriend:     res.IsFriend,
		},
	}, nil
}

func (cs *ConduitService) CanAddFriend(ctx context.Context, req *v1.CanAddFriendReq) (*v1.CanAddFriendRes, error) {
	res, err := cs.pc.CanAddFriend(ctx, req.TargetId)
	fmt.Println(res)
	if err != nil {
		log.Printf("CanAddFriend err: %v", err)

		return &v1.CanAddFriendRes{
			Code: 1,
			Res:  ErrorToRes(err),
			//Data: nil,
		}, nil
	}

	return &v1.CanAddFriendRes{
		Code: 0,
		Res:  ErrorToRes(err),
		//Data: res,
	}, nil
}
