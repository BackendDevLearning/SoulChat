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

func (cs *ConduitService) LoadMyGroup(ctx context.Context, req *v1.LoadMyGroupRequest) (*v1.LoadMyGroupReply, error) {
	info, err := cs.gc.LoadMyGroup(ctx, req.UserId)
	if err != nil {
		return &v1.LoadMyGroupReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}
	return &v1.LoadMyGroupReply{
		Code:    0,
		Res:     nil,
		Data:    info
	}, nil
}


func (cs *ConduitService) LoadJoinGroup(ctx context.Context, req *v1.LoadJoinGroupRequest) (*v1.LoadJoinGroupReply, error) {
	info, err := cs.gc.LoadJoinGroup(ctx, req.UserId)
	if err != nil {
		return &v1.LoadJoinGroupReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}
	return &v1.LoadJoinGroupReply{
		Code: 0,
		Res:  nil,
		Data: info,
	}, nil
}

func (cs *ConduitService) SetAdmin(ctx context.Context, req *v1.SetAdminRequest) (*v1.SetAdminReply, error) {
	err := cs.gc.SetAdmin(ctx, req.UserId, req.GroupId)
	if err != nil {
		return &v1.SetAdminReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}
	return &v1.SetAdminReply{
		Code: 0,
		Res:  nil,
	}, nil
}