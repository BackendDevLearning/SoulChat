package user

import (
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginRepo interface {
	GetUserByPhone(ctx context.Context, email string) (*UserLoginReply, error)

	//CreateUser(ctx context.Context, login *UserLogin) error
	//UpdateByCache(user *UserLogin) error
	//Load(user *UserLogin) error
}

type UserLoginCase struct {
	ul   UserLoginRepo
	jwtc *conf.JWT

	log *log.Helper
}

func NewUserLoginCase(ul UserLoginRepo, jwtc *conf.JWT, logger log.Logger) *UserLoginCase {
	return &UserLoginCase{
		ul:   ul,
		jwtc: jwtc,
		log:  log.NewHelper(logger),
	}
}

type UserLoginReply struct {
	ID           uint
	Phone        string
	Username     string
	Token        string
	Bio          string
	Image        string
	PasswordHash string
}

func (ul *UserLoginCase) Login(ctx context.Context, phone, password string) (*UserLoginReply, error) {
	if len(phone) == 0 {
		return nil, errors.New(422, "email", "cannot be empty")
	}
	u, err := ul.ul.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if !verifyPassword(u.PasswordHash, password) {
		return nil, errors.Unauthorized("user", "login failed")
	}

	return &UserLoginReply{
		Phone:    u.Phone,
		Username: u.Username,
		Bio:      u.Bio,
		Image:    u.Image,
		Token:    ul.generateToken1(u.ID),
	}, nil
}

func verifyPassword(hashed, input string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(input)); err != nil {
		return false
	}
	return true
}

func (ul *UserLoginCase) generateToken1(userID uint) string {
	return auth.GenerateToken(ul.jwtc.Secret, userID)
}
