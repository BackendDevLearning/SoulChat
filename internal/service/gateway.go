package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"
)

func (cs *ConduitService) Login(ctx context.Context, req *v1.LoginRequest) (reply *v1.LoginReply, err error) {
	u, err := cs.gc.Login(ctx, req.User.Phone, req.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		User: &v1.LoginReply_User{
			Code:  200,
			Res:   "success",
			Token: "123456",
		},
	}, nil
}

func (cs *ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (reply *v1.RegisterReply, err error) {
	u, err := cs.gc.Register(ctx, req.User.Username, req.User.Phone, req.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		User: &v1.RegisterReply_User{
			Code:  u.Code,
			Res:   u.Res,
			Token: u.Token,
		},
	}, nil
}
