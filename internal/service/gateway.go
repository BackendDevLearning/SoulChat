package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"
)

func (s *ConduitService) Login(ctx context.Context, req *v1.LoginRequest) (reply *v1.LoginResponse, err error) {
	u, err := s.ur.Login(ctx, req.User.Phone, req.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.UserReply{
		code: 0,
		res:  "success",
	}, nil
}

func (s *ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (reply *v1.RegisterResponse, err error) {
	u, err := s.ur.Register(ctx, req.User.Username, req.User.Phone, req.User.Password)
	if err != nil {
		return nil, err
	}
	return &v1.UserReply{
		User: &v1.UserReply_User{
			Phone:    u.Phone,
			Username: u.Username,
			Token:    u.Token,
		},
	}, nil
}
