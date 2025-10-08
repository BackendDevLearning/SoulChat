package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"
	"kratos-realworld/internal/common"
	"log"
)

func ConvertToMessage(req *v1.GetMessagesRequest) *common.MessageRequest {

	return &common.MessageRequest{
		MessageType:    req.MessageType,
		Uuid:           req.Uuid,
		FriendUsername: req.FriendUsername,
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
