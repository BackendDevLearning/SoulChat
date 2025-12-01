package data

import (
	"context"
	"errors"
	"fmt"
	bizGroup "kratos-realworld/internal/biz/messageGroup"
	"kratos-realworld/internal/model"
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

func (r *GroupInfoRepo) CreateGroup(group *bizGroup.GroupTB) error {
	rv := r.data.DB().Create(group)
	if rv.Error != nil {
		return rv.Error
	}

	// TODO: 缓存处理

	return nil
}