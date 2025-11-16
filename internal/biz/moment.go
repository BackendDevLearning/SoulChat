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
		UserID:   moment.UserID,
		MomentID: moment.ID,
		Action:   "create",
		PushIDs:  moment.PushIDs,
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
		UserID:   moment.UserID,
		MomentID: moment.ID,
		Action:   "delete",
		PushIDs:  moment.PushIDs,
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

func (uc *MomentUsecase) GetMoment(ctx context.Context, momentID uint32) (*bizMoment.MomentTB, error) {
	moment, err := uc.mr.GetMoment(ctx, momentID)
	if err != nil {
		uc.log.Errorf("GetMoment error: %v", err)
		return nil, NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Get moment failed")
	}
	return moment, nil
}

func (uc *MomentUsecase) GetMomentMeta(ctx context.Context, momentID uint32) (*bizMoment.MomentsMetaTB, error) {
	momentMeta, err := uc.mr.GetMomentMeta(ctx, momentID)
	if err != nil {
		uc.log.Errorf("GetMomentMeta error: %v", err)
		return nil, NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Get moment meta failed")
	}
	return momentMeta, nil
}

func (uc *MomentUsecase) CreateComment(ctx context.Context, comment *bizMoment.CommentsTB) error {
	err := uc.mr.CreateComment(ctx, comment)
	if err != nil {
		uc.log.Errorf("CreateComment error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Create comment failed")
	}
	return nil
}

func (uc *MomentUsecase) DeleteComment(ctx context.Context, commentID uint32) error {
	err := uc.mr.DeleteComment(ctx, commentID)
	if err != nil {
		uc.log.Errorf("DeleteComment error: %v", err)
		return NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Delete comment failed")
	}
	return nil
}

func (uc *MomentUsecase) UpdateMomentBlackList(ctx context.Context, momentID uint32, newBlackListIDs []uint32) ([]uint32, []uint32, error) {
	addedIDs, removedIDs, err := uc.mr.UpdateMomentBlackList(ctx, momentID, newBlackListIDs)
	if err != nil {
		uc.log.Errorf("UpdateMomentBlackList error: %v", err)
		return nil, nil, NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Update moment black list failed")
	}
	// 发送动态黑名单更新消息到 Kafka
	msg := &kafka.MomentMessage{
		MomentID: momentID,
		Action:   "blacklist_update_added",
		PushIDs:  addedIDs,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		uc.log.Errorf("Marshal MomentMessage error: %v", err)
		return nil, nil, NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Marshal moment message failed")
	}
	kafka.Send(body)
	uc.log.Infof("Send MomentMessage to Kafka: %s", body)
	// 发送动态黑名单更新消息到 Kafka
	msg = &kafka.MomentMessage{
		MomentID: momentID,
		Action:   "blacklist_update_removed",
		PushIDs:  removedIDs,
	}
	body, err = json.Marshal(msg)
	if err != nil {
		uc.log.Errorf("Marshal MomentMessage error: %v", err)
		return nil, nil, NewErr(ErrCodeMomentFailed, MOMENT_FAILED, "Marshal moment message failed")
	}
	kafka.Send(body)
	uc.log.Infof("Send MomentMessage to Kafka: %s", body)
	return addedIDs, removedIDs, nil
}
