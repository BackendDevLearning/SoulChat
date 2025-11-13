package sms

import (
	"context"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"kratos-realworld/internal/conf"
	"math"
	"math/rand"
	"strconv"
	"sync"
)

type SmsService struct {
	conf       *conf.Sms
	smsClient  *dysmsapi.Client
	clientOnce sync.Once
	clientErr  error
}

func NewSmsService(smsConf *conf.Sms) *SmsService {
	return &SmsService{
		conf: smsConf,
	}
}

// createClient 创建短信客户端（懒加载+单例）
func (s *SmsService) createClient() (*dysmsapi.Client, error) {
	s.clientOnce.Do(func() {
		config := &openapi.Config{
			AccessKeyId:     tea.String(s.conf.AccessKeyId),
			AccessKeySecret: tea.String(s.conf.AccessKeySecret),
			Endpoint:        tea.String(s.conf.Endpoint), // 从配置读取
		}
		s.smsClient, s.clientErr = dysmsapi.NewClient(config)
	})
	return s.smsClient, s.clientErr
}

// 生成指定位数的随机验证码
func (s *SmsService) generateCode(num int32) int {
	return rand.Intn(9*int(math.Pow(10, float64(num-1)))) + int(math.Pow(10, float64(num-1)))
}

// 发送短信验证码
func (s *SmsService) SendCode(ctx context.Context, phone string) (string, error) {
	client, err := s.createClient()
	if err != nil {
		return "", fmt.Errorf("failed to create sms client: %v", err)
	}

	code := strconv.Itoa(s.generateCode(s.conf.VerificationCode.Length))

	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(s.conf.SignName),
		TemplateCode:  tea.String(s.conf.TemplateCode),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", code)),
	}

	_, err = client.SendSms(req)
	if err != nil {
		return "", fmt.Errorf("failed to send sms: %v", err)
	}

	return code, nil
}
