package biz

import (
	"context"
	"encoding/json"
	bizMoment "kratos-realworld/internal/biz/moments"
	kafka "kratos-realworld/internal/kafka"

	"github.com/go-kratos/kratos/v2/log"
)

type MomentUsecase struct {
	mr bizMoment.MomentRepo

	log *log.Helper
}

func NewMomentUsecase(mr bizMoment.MomentRepo, logger log.Logger) *MomentUsecase {
	return &MomentUsecase{
		mr:  mr,
		log: log.NewHelper(logger),
	}
}

func (uc *MomentUsecase) CreateMoment(ctx context.Context, moment *bizMoment.MomentTB) error {
	err := uc.mr.CreateMoment(ctx, moment)
	if err != nil {
		uc.log.Errorf("CreateMoment error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Create moment failed")
	}

	// 发送动态创建消息到 Kafka
	msg := &kafka.MomentMessage{
		UserID:        moment.UserID,
		MomentID:      moment.ID,
		Action:        "create",
		ReceiveBoxIDs: moment.ReceiveBoxIDs,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		uc.log.Errorf("Marshal MomentMessage error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Marshal moment message failed")
	}
	kafka.Send(body)
	uc.log.Infof("Send MomentMessage to Kafka: %s", body)

	return nil
}

func (uc *MomentUsecase) DeleteMoment(ctx context.Context, momentID uint32) error {
	moment, err := uc.mr.DeleteMoment(ctx, momentID)
	if err != nil {
		uc.log.Errorf("DeleteMoment error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Delete moment failed")
	}

	// 发送动态删除消息到 Kafka
	msg := &kafka.MomentMessage{
		UserID:        moment.UserID,
		MomentID:      moment.ID,
		Action:        "delete",
		ReceiveBoxIDs: moment.ReceiveBoxIDs,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		uc.log.Errorf("Marshal MomentMessage error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Marshal moment message failed")
	}
	kafka.Send(body)
	uc.log.Infof("Send MomentMessage to Kafka: %s", body)

	return nil
}
