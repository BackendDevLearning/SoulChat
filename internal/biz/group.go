package biz

import (
	bizGroup "kratos-realworld/internal/biz/messageGroup"
	"context"
	"fmt"
	"kratos-realworld/internal/common"
	bizGroup "kratos-realworld/internal/biz/messageGroup"
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

	err := gc.gir.CreateGroup(group)
	if err != nil {
		return 0, err
	}
	return group.Uuid, nil
}