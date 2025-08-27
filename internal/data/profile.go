package data

import (
	"context"
	bizProfile "kratos-realworld/internal/biz/profile"
	"kratos-realworld/internal/model"

	"github.com/go-kratos/kratos/v2/log"
)

type ProfileRepo struct {
	data *model.Data
	log  *log.Helper
}

func NewProfileRepo(data *model.Data, logger log.Logger) bizProfile.ProfileRepo {
	return &ProfileRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *ProfileRepo) CreateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	return nil
}

func (r *ProfileRepo) GetProfileByUserID(ctx context.Context, userID uint) (*bizProfile.ProfileTB, error) {
	return nil, nil
}

func (r *ProfileRepo) UpdateProfile(ctx context.Context, profile *bizProfile.ProfileTB) error {
	return nil
}

func (r *ProfileRepo) IncrementFollowCount(ctx context.Context, userID uint, delta int) error {
	return nil
}

func (r *ProfileRepo) IncrementFanCount(ctx context.Context, userID uint, delta int) error {
	return nil
}
