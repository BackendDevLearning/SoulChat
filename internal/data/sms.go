package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	bizUser "kratos-realworld/internal/biz/user"
	"kratos-realworld/internal/model"
	"kratos-realworld/internal/pkg/middleware/sms"
)

type smsRepo struct {
	data       *model.Data
	log        *log.Helper
	smsService *sms.SmsService
}

func NewSmsRepo(data *model.Data, logger log.Logger, smsService *sms.SmsService) bizUser.SmsRepo {
	return &smsRepo{
		data:       data,
		log:        log.NewHelper(logger),
		smsService: smsService,
	}
}

func (r *smsRepo) SendCode(ctx context.Context, phone string) (string, error) {
	code, err := r.smsService.SendCode(ctx, phone)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (r *smsRepo) SaveCode(ctx context.Context, phone, code string) error {
	redisKey := UserRedisKey(UserSMSPrefix, "Phone", phone)

	// 如果是第一次发验证码，就直接写入redis缓存，不是第一次，会默认覆盖掉旧的验证码，并重置TTL
	if _, err := r.data.Cache().HSet(ctx, redisKey, "Code", code); err != nil {
		r.log.Warnf("failed to write SMS verification code to cache: %v", err)
	} else {
		r.data.Cache().Expire(ctx, redisKey, UserSMSTTL)
		r.log.Debugf("SMS verification code cached successfully, set TTL to %s for key %s", UserSMSTTL, redisKey)
	}

	return nil
}

func (r *smsRepo) VerifyCode(ctx context.Context, phone string, inputCode string) (bool, error) {
	redisKey := UserRedisKey(UserSMSPrefix, "Phone", phone)

	code, err := r.data.Cache().HGet(ctx, redisKey, "Code")
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.log.Infof("verification code expired or not found for phone: %s", phone)
			return false, nil // 不视为系统错误，而是验证码无效
		}
		r.log.Errorf("failed to get SMS verification code from cache for %s: %v", phone, err)
		return false, err
	}

	if code != inputCode {
		return false, nil
	}
	return true, nil
}
