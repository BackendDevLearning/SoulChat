package biz

import "kratos-realworld/internal/biz/user"

type GateWay struct {
	couponRepo user.UserLoginRepo
}

func NewGatWayCase(pr user.UserLoginRepo) *GateWay {
	return &GateWay{
		couponRepo: pr,
	}
}
