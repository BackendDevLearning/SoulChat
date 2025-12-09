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
	group := &bizGroup.GroupTB{
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
	if err := r.data.Cache().DelKeysWithPattern("group_mygroup_list_" + fmt.Sprintf("%d", group.CreaterID)); err != nil {
		r.log.Warnf("CreateGroup delete cache err: %v\n", err)
		// 缓存删除失败不影响主流程，只记录警告
	}

	return nil
}

func (r *GroupInfoRepo) AddGroupMember(user_id uint32, group_id uint32) error {
	groupMember := &bizGroup.GroupMemberTB{
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
	cacheKey := fmt.Sprintf("group_mygroup_list_%d", UserId)
	rspString, err := r.data.Cache().GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			var groupsList []*bizGroup.GroupTB
			// 缓存未命中，从数据库查询
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
				r.log.Warnf("LoadMyGroup set cache err: %v\n", err)
				// 缓存设置失败不影响主流程，返回查询结果
			}
		} else {
			r.log.Errorf("LoadMyGroup err: %v\n", err)
			return nil, NewErr(ErrCodeDBQueryFailed, DB_QUERY_FAILED, "failed to query my group")
		}
			
	}
	
	// 从缓存中解析数据
	var groups []res.LoadMyGroupData
	err = json.Unmarshal([]byte(rspString), &groups)
	if err != nil {
		r.log.Errorf("LoadMyGroup unmarshal err: %v\n", err)
		return nil, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to unmarshal group data")
	}
	return groups, nil
}

func (r *GroupInfoRepo) LoadJoinGroup(UserId uint32) ([]res.LoadMyGroupData, error) {
	// 先查redis
	cacheKey := fmt.Sprintf("group_joingroup_list_%d", UserId)
	rspString, err := r.data.Cache().GetKeyNilIsErr(cacheKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 缓存未命中，从数据库查询
			var groupsList []*bizGroup.GroupTB
			
			// 使用联表查询：通过 GroupMember 表关联查询用户加入的所有群组
			err := r.data.DB().
				Table("t_group g").
				Select("g.id as group_id, g.name, g.avatar").
				Joins("INNER JOIN t_groupMember gm ON g.id = gm.group_id").
				Where("gm.user_id = ? AND gm.deleted_at IS NULL AND g.deleted_at IS NULL", UserId).
				Scan(&groupsList).Error
			
			if err != nil {
				r.log.Errorf("LoadJoinGroup query err: %v\n", err)
				return nil, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to query join group")
			}

			for _, group := range groupsList {
				groups = append(groups, res.LoadMyGroupData{
					GroupId: group.ID,
					Name:    group.Name,
					Avatar:  group.Avatar,
				})
			}

			// 将结果缓存到 Redis
			rspString, err := json.Marshal(groups)
			if err != nil {
				r.log.Errorf("LoadJoinGroup marshal err: %v\n", err)
				return nil, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to marshal group data")
			}
			
			ctx := context.Background()
			err = r.data.Cache().SetKey(ctx, cacheKey, string(rspString), common.GroupMyGroupListCacheTTL)
			if err != nil {
				r.log.Warnf("LoadJoinGroup set cache err: %v\n", err)
				// 缓存设置失败不影响主流程，返回查询结果
			}
			
			return groups, nil
		} else {
			r.log.Errorf("LoadJoinGroup cache err: %v\n", err)
			return nil, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to get cache")
		}
	}
	
	// 从缓存中解析数据
	var groups []res.LoadMyGroupData
	err = json.Unmarshal([]byte(rspString), &groups)
	if err != nil {
		r.log.Errorf("LoadJoinGroup unmarshal err: %v\n", err)
		return nil, biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to unmarshal group data")
	}
	return groups, nil
}

func (r *GroupInfoRepo) SetAdmin(UserId uint32, GroupId uint32, CallerId uint32) error {
	// 1. 验证群组是否存在
	var group bizGroup.GroupTB
	err := r.data.DB().Where("id = ? AND deleted_at IS NULL", GroupId).First(&group).Error
	if err != nil {
		r.log.Errorf("SetAdmin: group not found, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "group not found")
	}

	// 验证CallerId是否是群主
	if group.CreaterID != CallerId {
		r.log.Errorf("SetAdmin: caller is not the group creator, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "caller is not the group creator")
	}

	// 2. 验证用户是否是群成员
	var groupMember bizGroup.GroupMemberTB
	err = r.data.DB().Where("user_id = ? AND group_id = ? AND deleted_at IS NULL", UserId, GroupId).First(&groupMember).Error
	if err != nil {
		r.log.Errorf("SetAdmin: user is not a group member, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "user is not a group member")
	}

	// 3. 检查用户是否已经是管理员
	if groupMember.Role == common.GroupAdmin {
		r.log.Warnf("SetAdmin: user %d is already an admin of group %d\n", UserId, GroupId)
		// 已经是管理员，直接返回成功
		return nil
	}

	// 4. 更新 GroupMemberTB 表中的 Role 字段为管理员
	err = r.data.DB().Model(&groupMember).Update("role", common.GroupAdmin).Error
	if err != nil {
		r.log.Errorf("SetAdmin: failed to update group member role, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to update group member role")
	}

	// 5. 更新 GroupTB 表中的 Adminer 字段，将用户ID添加到管理员列表中
	var adminList []uint32
	if group.Adminer != "" {
		err = json.Unmarshal([]byte(group.Adminer), &adminList)
		if err != nil {
			r.log.Errorf("SetAdmin: failed to unmarshal adminer list, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to parse adminer list")
		}
	}

	// 检查用户ID是否已经在管理员列表中
	alreadyAdmin := false
	for _, adminId := range adminList {
		if adminId == UserId {
			alreadyAdmin = true
			break
		}
	}

	// 如果不在列表中，则添加
	if !alreadyAdmin {
		adminList = append(adminList, UserId)
		adminListBytes, err := json.Marshal(adminList)
		if err != nil {
			r.log.Errorf("SetAdmin: failed to marshal adminer list, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to marshal adminer list")
		}

		err = r.data.DB().Model(&group).Update("adminer", string(adminListBytes)).Error
		if err != nil {
			r.log.Errorf("SetAdmin: failed to update group adminer, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to update group adminer")
		}
	}

	return nil
}

func (r *GroupInfoRepo) RemoveAdmin(UserId uint32, GroupId uint32, CallerId uint32) error {
	// 1. 验证群组是否存在
	var group bizGroup.GroupTB
	err := r.data.DB().Where("id = ? AND deleted_at IS NULL", GroupId).First(&group).Error
	if err != nil {
		r.log.Errorf("RemoveAdmin: group not found, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "group not found")
	}

	// 验证CallerId是否是群主
	if group.CreaterID != CallerId {
		r.log.Errorf("RemoveAdmin: caller is not the group creator, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "caller is not the group creator")
	}

	// 2. 验证用户是否是管理员
	var groupMember bizGroup.GroupMemberTB
	err = r.data.DB().Where("user_id = ? AND group_id = ? AND deleted_at IS NULL", UserId, GroupId).First(&groupMember).Error
	if err != nil {
		r.log.Errorf("RemoveAdmin: user is not a group member, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "user is not a group member")
	}

	// 移除用户的管理员
	err = r.data.DB().Model(&groupMember).Update("role", common.GroupMember).Error
	if err != nil {
		r.log.Errorf("RemoveAdmin: failed to update group member role, err: %v\n", err)
		return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to update group member role")
	}

	// 更新 GroupTB 表中的 Adminer 字段，将用户ID从管理员列表中移除
	var adminList []uint32
	if group.Adminer != "" {
		err = json.Unmarshal([]byte(group.Adminer), &adminList)
		if err != nil {
			r.log.Errorf("RemoveAdmin: failed to unmarshal adminer list, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to parse adminer list")
		}
		for i, adminId := range adminList {
			if adminId == UserId {
				adminList = append(adminList[:i], adminList[i+1:]...)
				break
			}
		}
		adminListBytes, err := json.Marshal(adminList)
		if err != nil {
			r.log.Errorf("RemoveAdmin: failed to marshal adminer list, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to marshal adminer list")
		}
		err = r.data.DB().Model(&group).Update("adminer", string(adminListBytes)).Error
		if err != nil {
			r.log.Errorf("RemoveAdmin: failed to update group adminer, err: %v\n", err)
			return biz.NewErr(biz.ErrCodeDBQueryFailed, biz.DB_QUERY_FAILED, "failed to update group adminer")
		}
	}

	return nil
}