package service

import (
	"context"
	v1 "kratos-realworld/api/conduit/v1"

	"log"
)

func (cs *ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	res, err := cs.gt.Register(ctx, req.Username, req.Phone, req.Password)
	if err != nil {
		log.Printf("Register error! %v", err)

		return &v1.RegisterReply{
			Code:  1,
			Res:   ErrorToRes(err),
			Token: "",
		}, nil
	}

	return &v1.RegisterReply{
		Code:  0,
		Res:   ErrorToRes(err),
		Token: res.Token,
	}, nil
}

func (cs *ConduitService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	res, err := cs.gt.Login(ctx, req.Phone, req.Password)

	if err != nil {
		log.Printf("Login error! %v", err)

		return &v1.LoginReply{
			Code:  1,
			Res:   ErrorToRes(err),
			Token: "",
		}, nil
	}

	return &v1.LoginReply{
		Code:  0,
		Res:   ErrorToRes(err),
		Token: res.Token,
	}, nil

	//res := &v1.LoginReply{
	//	Code:  0,
	//	Res:   "success",
	//	Token: "",
	//}

	//session, err = cs.gt.Login(ctx, req.Phone, req.Password)

	//if err != nil {
	//	res.Code = 1
	//	res.Res = "servce Login error"
	//}

	// 登陆成功，设置cookie
	//ctx.SetCookie(SessionKey, session, CookieExpire, "/", "", false, true)
}
