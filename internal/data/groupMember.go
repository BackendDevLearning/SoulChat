package data

import (
	"context"
	"errors"
	"fmt"
	bizGroup "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/model"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"encoding/json"
	"kratos-realworld/internal/common"
	"kratos-realworld/internal/common/res"
	"kratos-realworld/internal/biz"
	"gorm.io/gorm"
)

type GroupMemberRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewGroupMemberRepo(data *model.Data, logger log.Logger) bizGroup.GroupMemberRepo {
	return &GroupMemberRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *GroupMemberRepo) AddGroupMember(user_id uint32, group_id uint32) error {
	groupMember := &bizGroup.GroupMemberTB{
		UserID: user_id,
		GroupID: group_id,
		Nickname: "",
		Mute: 0,
		Role: common.GroupMember,
	}
	rv := r.data.DB().Create(groupMember)
	if rv.Error != nil {
		r.log.Errorf("AddGroupMember err: %v\n", rv.Error)
		return rv.Error
	}


	// 先查redis
	cacheKey := fmt.Sprintf("group_mygroup_list_%d", user_id)
	rspString, err := r.data.Cache().GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 缓存未命中
			r.log.Errorf("AddGroupMember GetKeyNilIsErr err: %v\n", err)
		}
	} else {
		// 缓存命中
		r.log.Infof("AddGroupMember GetKeyNilIsErr err: %v\n", err)
		// 删除缓存
		err = r.data.Cache().DelKey(cacheKey)
		if err != nil {
			r.log.Errorf("AddGroupMember DelKey err: %v\n", err)
		}
	}
	return nil
}


func (r *GroupMemberRepo) IsUserInGroup(UserId uint32, GroupId uint32) (bool, error) {
	var groupMember bizGroup.GroupMemberTB
	err := r.data.DB().Where("user_id = ? AND group_id = ? AND deleted_at IS NULL", UserId, GroupId).First(&groupMember).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，说明用户不在群组中，返回 false
			return false, nil
		}
		r.log.Errorf("IsUserInGroup: query failed, err: %v\n", err)
		return false, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to check user in group")
	}
	return true, nil
}

func (r *GroupMemberRepo) IsUserAdmin(UserId uint32, GroupId uint32) (bool, error) {
	var groupMember bizGroup.GroupMemberTB
	err := r.data.DB().Where("user_id = ? AND group_id = ? AND deleted_at IS NULL AND role = ?", UserId, GroupId, common.GroupAdmin).First(&groupMember).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，说明用户不是管理员，返回 false
			return false, nil
		}
		r.log.Errorf("IsUserAdmin: query failed, err: %v\n", err)
		return false, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to check user admin")
	}
	return true, nil
}