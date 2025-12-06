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
)

type GroupInfoRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewGroupInfoRepo(data *model.Data, logger log.Logger) bizGroup.GroupInfoRepo {
	return &GroupInfoRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *GroupInfoRepo) CreateGroup(user_id uint32, name string, mode uint32, add_mode uint32, intro string) error {
	group = &bizGroup.GroupTB{
		Uuid:      fmt.Sprintf("G%s", common.GetNowAndLenRandomString(11)),
		CreaterID: user_id,
		Name:      name,
		Mode:      mode,
		AddMode:   add_mode,
		Intro:     intro,
		Member:    fmt.Sprintf("[%d]", user_id),
		Adminer:   fmt.Sprintf("[%d]", user_id),
		Avatar:    common.GetDefaultGroupAvatar(),
		MemberCount: 1,
		Notice:    "",
	}
	rv := r.data.DB().Create(group)
	if rv.Error != nil {
		return rv.Error
	}

	// 删除缓存，因为创建群后，群列表需要更新
	if err := r.data.Cache().DelKeysWithPattern("group_mygroup_list_" + group.CreaterID); err != nil {
		r.log.Errorf("CreateGroup err: %v\n", err)
		return NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to create group")
	}

	return nil
}

func (r *GroupInfoRepo) AddGroupMember(user_id uint32, group_id uint32) error {
	groupMember = &bizGroup.GroupMemberTB{
		UserID: user_id,
		GroupID: group_id,
		Nickname: "",
		Mute: 0,
		Role: common.GroupAdmin,
	}
	rv := r.data.DB().Create(groupMember)
	if rv.Error != nil {
		r.log.Errorf("AddGroupMember err: %v\n", rv.Error)
		return rv.Error
	}
	return nil
}

func (r *GroupInfoRepo) LoadMyGroup(UserId uint32) ([]res.LoadMyGroupData, error) {
    // 先查redis
	rspString, err := r.data.Cache().GetKeyNilIsErr("group_mygroup_list_" + UserId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var groupsList []*bizGroup.GroupTB
			var groups []res.LoadMyGroupData
			err := r.data.DB().Where("CreaterID = ?", UserId).Find(&groupsList).Error
			if err != nil {
				r.log.Errorf("LoadMyGroup err: %v\n", err)
				return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
			}

			for _, group := range groupsList {
				groups = append(groups, res.LoadMyGroupData{
					GroupId: group.ID,
					Name:    group.Name,
					Avatar:  group.Avatar,
				})
			}
			rspString, err := json.Marshal(groups)
			if err != nil {
				r.log.Errorf("LoadMyGroup err: %v\n", err)
				return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
			}
			err = r.data.Cache().SetKey(ctx, "group_mygroup_list_"+UserId, string(rspString), common.GroupMyGroupListCacheTTL)
			if err != nil {
				r.log.Errorf("LoadMyGroup err: %v\n", err)
				return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
			}
		} else {
			r.log.Errorf("LoadMyGroup err: %v\n", err)
			return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
		}
			
	}
	var groups []res.LoadMyGroupData
	err = json.Unmarshal([]byte(rspString), &groups)
	if err != nil {
		r.log.Errorf("LoadMyGroup err: %v\n", err)
		return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
	}
	return groups, nil
}