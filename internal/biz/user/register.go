package user

import (
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/pkg/middleware/auth"

	"context"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserRegisterTB struct {
	ID           uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserName     string     `gorm:"column:UserName;type:varchar(50);comment:账号;NOT NULL" json:"UserName"`
	Phone        string     `gorm:"column:Phone;type:varchar(20);comment:手机号码;NOT NULL" json:"Phone"`
	PasswordHash string     `gorm:"column:PassWord;type:text;comment:密码;NOT NULL" json:"PassWord"`
	Token        string     `gorm:"column:Token;type:varchar(50);comment:Token" json:"Token"`
	Bio          string     `gorm:"column:Bio;type:text;comment:简介" json:"Bio"`
	Image        string     `gorm:"column:Image;type:varchar(255);comment:头像链接" json:"Image"`
	SysCreated   *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated   *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

func (m *UserRegisterTB) TableName() string {
	return "t_user_register"
}

type UserRegisterRepo interface {
	CreateUser(ctx context.Context, userRegister *UserRegisterTB) error

	//CreateUser(ctx context.Context, login *UserRegister) error
	//UpdateByCache(user *UserRegister) error
	//Load(user *UserRegister) error
}

type UserRegisterCase struct {
	ur   UserRegisterRepo
	jwtc *conf.JWT

	log *log.Helper
}

func NewUserRegisterCase(ur UserRegisterRepo, jwtc *conf.JWT, logger log.Logger) *UserRegisterCase {
	return &UserRegisterCase{
		ur:   ur,
		jwtc: jwtc,
		log:  log.NewHelper(logger),
	}
}

type UserRegisterReply struct {
	Phone    string
	Username string
	Token    string
	Bio      string
	Image    string
}

func hashPassword(pwd string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (ur *UserRegisterCase) Register(ctx context.Context, username, phone, password string) (*UserRegisterReply, error) {
	userRegister := &UserRegisterTB{
		Phone:        phone,
		UserName:     username,
		PasswordHash: hashPassword(password),
	}
	if err := ur.ur.CreateUser(ctx, userRegister); err != nil {
		return nil, err
	}
	return &UserRegisterReply{
		Phone:    phone,
		Username: username,
		Token:    ur.generateToken(userRegister.ID),
	}, nil
}

func (ur *UserRegisterCase) generateToken(userID uint) string {
	return auth.GenerateToken(ur.jwtc.Secret, userID)
}
