package service

import (
	"context"

	v1 "kratos-realworld/api/conduit/v1"
)

func (cs *ConduitService) CreateGroup(ctx context.Context, req *v1.CreateGroupRequest) (*v1.CreateGroupReply, error) {
	groupId, err := cs.gc.CreateGroup(ctx, req.user_id, req.name, req.mode, req.add_mode, req.intro)
	if err != nil {
		return &v1.CreateGroupReply{
			Code: 1,
			Res:  ErrorToRes(err),
			group_id: 0,
		}, nil
	}
	return &v1.CreateGroupReply{
		Code: 0,
		Res:  nil,
		group_id: groupId,
	}, nil
}