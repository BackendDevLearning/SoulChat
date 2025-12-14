package biz

import (
	bizGroup "kratos-realworld/internal/biz/messageGroup"
	"context"
	"fmt"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/common/res"
	"github.com/go-kratos/kratos/v2/log"
	"errors"
)

type GroupUseCase struct {
	gir bizGroup.GroupInfoRepo
	gmr bizGroup.GroupMemberRepo
	mr  bizGroup.MessageRepo
	log *log.Helper
}

func (gc *GroupUseCase) NewGroupUseCase(gir bizGroup.GroupInfoRepo, gmr bizGroup.GroupMemberRepo, mr bizGroup.MessageRepo, logger log.Logger) *GroupUseCase {
	return &GroupUseCase{
		gir: gir,
		gmr: gmr,
		mr:  mr,
		log: log.NewHelper(logger),
	}
}

func (gc *GroupUseCase) CreateGroup(ctx context.Context, user_id uint32, name string, mode uint32, add_mode uint32, intro string) (uint32, error) {
	// 创建群组，获取群组ID
	group_id, err := gc.gir.CreateGroup(user_id, name, mode, add_mode, intro)
	if err != nil {
		gc.log.Errorf("CreateGroup err: %v\n", err)
		return 0, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to create group")
	}
	
	// 创建群组后，添加创建者为群成员
	err = gc.gmr.AddGroupMember(user_id, group_id)
	if err != nil {
		gc.log.Errorf("AddGroupMember err: %v\n", err)
		return 0, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to add group member")
	}
	
	return group_id, nil
}

func (gc *GroupUseCase) LoadMyGroup(ctx context.Context, UserId uint32) ([]res.LoadMyGroupData, error) {
	groups, err := gc.gir.LoadMyGroup(UserId)
	if err != nil {
		gc.log.Errorf("LoadMyGroup err: %v\n", err)
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
	}
	
	if len(groups) == 0 {
		// 返回空数组而不是错误
		return []res.LoadMyGroupData{}, nil
	}
	
	return groups, nil
}

func (gc *GroupUseCase) LoadJoinGroup(ctx context.Context, UserId uint32) ([]res.LoadMyGroupData, error) {
	groups, err := gc.gir.LoadJoinGroup(UserId)
	if err != nil {
		gc.log.Errorf("LoadJoinGroup err: %v\n", err)
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query join group")
	}
	
	if len(groups) == 0 {
		// 返回空数组而不是错误
		return []res.LoadMyGroupData{}, nil
	}
	
	return groups, nil
}

func (gc *GroupUseCase) SetAdmin(ctx context.Context, UserId uint32, GroupId uint32, CallerId uint32) error {
	err := gc.gir.SetAdmin(UserId, GroupId, CallerId)
	if err != nil {
		gc.log.Errorf("SetAdmin err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to set admin")
	}
	return nil
}

func (gc *GroupUseCase) RemoveAdmin(ctx context.Context, UserId uint32, GroupId uint32, CallerId uint32) error {
	err := gc.gir.RemoveAdmin(UserId, GroupId, CallerId)
	if err != nil {
		gc.log.Errorf("RemoveAdmin err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to remove admin")
	}
	return nil
}

func (gc *GroupUseCase) JoinGroup(ctx context.Context, UserId uint32, GroupId uint32) error {
	// 验证群组是否存在
	exists, err := gc.gir.IsGroupExist(GroupId)
	if err != nil {
		gc.log.Errorf("IsGroupExist err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to check group exist")
	}
	if !exists {
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "group not found")
	}
	
	// 验证用户是否已经在群组中
	userInGroup, err := gc.gmr.IsUserInGroup(UserId, GroupId)
	if err != nil {
		gc.log.Errorf("IsUserInGroup err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to check user in group")
	}
	if userInGroup {
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "user already in group")
	}
	
	
	// 添加用户为群成员
	err = gc.gmr.AddGroupMember(UserId, GroupId)
	if err != nil {
		gc.log.Errorf("JoinGroup AddGroupMember err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to join group")
	}

	// 更新群组成员列表
	err = gc.gir.AddMemberToJSON(GroupId, UserId)
	if err != nil {
		gc.log.Errorf("JoinGroup AddMemberToJSON err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to add member to group")
	}

	// 更新群组成员数量
	err = gc.gir.updateMemberCount(GroupId)
	if err != nil {
		gc.log.Errorf("updateMemberCount err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to update member count")
	}
	return nil
}