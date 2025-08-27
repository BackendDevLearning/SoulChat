package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"

	"log"
)

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
		Data: &v1.FollowFanReply_RelationData{
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
		Data: &v1.FollowFanReply_RelationData{
			SelfId:      res.SelfID,
			FollowCount: res.FollowCount,
			TargetId:    res.TargetID,
			FanCount:    res.FanCount,
		},
	}, nil
}
