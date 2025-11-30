package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/common"
	"github.com/go-kratos/kratos/v2/log"
	"kratos-realworld/internal/common/res"
)

func ConvertToMessage(req *v1.GetMessagesRequest) *common.MessageRequest {

	return &common.MessageRequest{
		MessageType: req.MessageType,
		Uuid:        req.Uuid,
		FriendUuid:  req.FriendUuid,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}
}

func (cs *ConduitService) GetMessages(ctx context.Context, req *v1.GetMessagesRequest) (*v1.GetMessagesReply, error) {
	err := cs.mc.GetMessages(ctx, *ConvertToMessage(req))
	if err != nil {
		log.Printf("GetMessages err: %v\n", err)

		return &v1.GetMessagesReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}

	return &v1.GetMessagesReply{
		Code: 0,
		Res:  ErrorToRes(err),
	}, nil
}

// response 的结构和messageResponse 的结构一致
func (cs *ConduitService) GetMessageList(ctx context.Context, req *v1.GetMessageListRequest) (*v1.GetMessageListReply, error) {
	res, err := cs.mc.GetMessageList(ctx, req.Uuid1, req.Uuid2)
	
	if err != nil {
		cs.log.Errorf("GetMessageList err: %v\n", err)
		return &v1.GetMessageListReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}
	
	return &v1.GetMessageListReply{
		Code: 0,
		Res:  ErrorToRes(err),
		Data: res,
	}, nil
}
