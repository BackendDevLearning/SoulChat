package service

import (
	"context"
	"fmt"
	v1 "kratos-realworld/api/conduit/v1"
	bizUser "kratos-realworld/internal/biz/user"
	"log"
)

func (cs *ConduitService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	res, err := cs.gt.Register(ctx, req.Username, req.Phone, req.Password)
	if err != nil {
		log.Printf("Register error: %v", err)

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
		log.Printf("Login error: %v", err)

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
}

func (cs *ConduitService) UpdateUserPassword(ctx context.Context, req *v1.UpdateUserPwdRequest) (*v1.UpdateUserPwdReply, error) {
	err := cs.gt.UpdateUserPassword(ctx, req.Phone, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Printf("UpdateUserPassword error: %v", err)

		return &v1.UpdateUserPwdReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}

	return &v1.UpdateUserPwdReply{
		Code: 0,
		Res:  ErrorToRes(err),
	}, nil
}

func ConvertToUpdateUserInfoFields(req *v1.UpdateUserInfoRequest) *bizUser.UpdateUserInfoFields {
	fmt.Println("Gender", req.Gender)
	fields := &bizUser.UpdateUserInfoFields{
		Username:   strPtr(req.Username),
		Bio:        strPtr(req.Bio),
		HeadImage:  strPtr(req.HeadImage),
		CoverImage: strPtr(req.CoverImage),
	}

	if req.Gender != v1.Gender_UNKNOWN {
		gender := uint32(req.Gender)
		fmt.Println("gender", gender)
		fields.Gender = &gender
	}

	if req.Birthday != nil {
		birthday := req.Birthday.AsTime()
		fields.Birthday = &birthday
	}

	return fields
}

func strPtr(s string) *string {
	if s == "" {
		return nil // 不更新该字段
	}
	return &s
}

func (cs *ConduitService) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoReply, error) {
	fields := ConvertToUpdateUserInfoFields(req)

	err := cs.gt.UpdateUserInfo(ctx, fields)

	if err != nil {
		log.Printf("UpdateUserInfo error: %v", err)

		return &v1.UpdateUserInfoReply{
			Code: 1,
			Res:  ErrorToRes(err),
		}, nil
	}

	return &v1.UpdateUserInfoReply{
		Code: 0,
		Res:  ErrorToRes(err),
	}, nil
}
