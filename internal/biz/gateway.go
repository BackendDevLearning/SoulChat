package biz

import "kratos-realworld/internal/biz/user"

type GateWay struct {
	couponRepo user.UserLogRepo
}

func NewGatWayCase(pr user.UserLogRepo) *GateWay {
	return &GateWay{
		couponRepo: pr,
	}
}
