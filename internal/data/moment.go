package data

import (
	"context"
	bizMoment "kratos-realworld/internal/biz/moments"
	"kratos-realworld/internal/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type MomentRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewMomentRepo(data *model.Data, log *log.Helper) *MomentRepo {
	return &MomentRepo{
		data: data,
		log:  log,
	}
}

func (r *MomentRepo) CreateMoment(ctx context.Context, moment *bizMoment.MomentTB) error {
	err := r.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(moment).Error
		if err != nil {
			return err
		}
		momentMeta := &bizMoment.MomentsMetaTB{
			ID:            moment.ID,
			UserID:        moment.UserID,
			MomentID:      moment.ID,
			Message:       moment.Message,
			MediaURL:      moment.MediaURL,
			ReceiveBoxIDs: moment.ReceiveBoxIDs,
		}
		err = tx.Create(momentMeta).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MomentRepo) GetMoment(ctx context.Context, momentID uint32) (*bizMoment.MomentTB, error) {
	var moment bizMoment.MomentTB
	err := r.data.DB().WithContext(ctx).Where("id = ?", momentID).First(&moment).Error
	if err != nil {
		return nil, err
	}
	return &moment, nil
}

func (r *MomentRepo) DeleteMoment(ctx context.Context, momentID uint32) (*bizMoment.MomentTB, error) {
	var moment *bizMoment.MomentTB
	moment, err := r.GetMoment(ctx, momentID)
	if err != nil {
		return nil, err
	}
	err = r.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&bizMoment.MomentTB{})
		//执行软删除
		err = tx.Where("id = ?", momentID).Set("deleted_at", time.Now()).Error
		if err != nil {
			return err
		}
		tx = tx.Model(&bizMoment.MomentsMetaTB{})
		err = tx.Where("moment_id = ?", momentID).Set("deleted_at", time.Now()).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return moment, nil
}

func (r *MomentRepo) CreateComment(ctx context.Context, comment *bizMoment.CommentsTB) error {
	err := r.data.DB().WithContext(ctx).Create(comment).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *MomentRepo) GetComments(ctx context.Context, momentID uint32) ([]*bizMoment.CommentsTB, error) {
	var comments []*bizMoment.CommentsTB
	err := r.data.DB().WithContext(ctx).Where("moment_id = ?", momentID).Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *MomentRepo) DeleteComment(ctx context.Context, commentID uint32) error {
	err := r.data.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tx = tx.Model(&bizMoment.CommentsTB{})
		//执行软删除
		err := tx.Where("id = ?", commentID).Set("deleted_at", time.Now()).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MomentRepo) GetMomentMeta(ctx context.Context, momentID uint32) (*bizMoment.MomentsMetaTB, error) {
	var momentMeta bizMoment.MomentsMetaTB
	err := r.data.DB().WithContext(ctx).Where("moment_id = ?", momentID).First(&momentMeta).Error
	if err != nil {
		return nil, err
	}
	return &momentMeta, nil
}
