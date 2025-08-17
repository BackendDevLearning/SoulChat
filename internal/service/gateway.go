package service

import (
	"context"
	"github.com/gin-gonic/gin"
	v1 "kratos-realworld/api/conduit/v1"
)

func (cs *ConduitService) Login(ctx *gin.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	res := &v1.LoginReply{
		Code:  0,
		Res:   "success",
		Token: "",
	}

	session, err := cs.gt.Login(ctx, req.Phone, req.Password)

	if err != nil {
		res.Code = 1
		res.Res = "servce Login error"
	}

	// 登陆成功，设置cookie
	ctx.SetCookie(SessionKey, session, CookieExpire, "/", "", false, true)

	return res, nil
}

func (cs *ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	res := &v1.RegisterReply{
		Code:  0,
		Res:   "success",
		Token: "",
	}

	result, err := cs.gt.Register(ctx, req.Username, req.Phone, req.Password)
	if err != nil {
		res.Code = 1
		res.Res = result
	}
	return res, nil
}
