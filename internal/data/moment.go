package data

import (
	"kratos-realworld/internal/model"

	"github.com/go-kratos/kratos/v2/log"
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
