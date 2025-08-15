package user

import (
	"context"
	"fmt"
	"kratos-realworld/internal/service"

	v1 "kratos-realworld/api/conduit/v1"
	//"kratos-realworld/internal/biz"
)

//type ConduitService struct {
//	*service.ConduitService
//}

func (s *service.ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (reply *v1.UserReply, err error) {
	fmt.Println("service Register")
	u, err := s.UR().Register(ctx, req.User.Username, req.User.Phone, req.User.Password)
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
