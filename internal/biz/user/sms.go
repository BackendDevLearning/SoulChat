package user

import "context"

// SMS服务封装到一个独立repo，这样写看起来层数比较复杂，但是后续如果考虑将阿里云更换为其它平台SDK，就只需要修改中间件
type SmsRepo interface {
	SendCode(ctx context.Context, phone string) (string, error)            // 调用第三方SMS服务
	SaveCode(ctx context.Context, phone, code string) error                // 将验证码缓存到redis
	VerifyCode(ctx context.Context, phone, inputCode string) (bool, error) // 从redis中取出验证码，并且和用户输入进行比较
}
