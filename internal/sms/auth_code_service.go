package sms

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"strconv"
	"sync"
	"time"
	
	"kratos-realworld/internal/conf"
	"kratos-realworld/internal/model"
)

type SmsService struct {
	conf        *conf.Sms
	smsClient   *dysmsapi20170525.Client
	clientOnce  sync.Once
	clientErr   error
}

func NewSmsService(smsConf *conf.Sms) *SmsService {
	return &SmsService{
		conf:  smsConf,
	}
}

// createClient 创建短信客户端（懒加载+单例）
func (s *SmsService) createClient() (*dysmsapi20170525.Client, error) {
	s.clientOnce.Do(func() {
		config := &openapi.Config{
			AccessKeyId:     tea.String(s.conf.AccessKeyId),
			AccessKeySecret: tea.String(s.conf.AccessKeySecret),
			Endpoint:        tea.String(s.conf.Endpoint), // 从配置读取
		}
		s.smsClient, s.clientErr = dysmsapi20170525.NewClient(config)
	})
	return s.smsClient, s.clientErr
}

func (s *SmsService) VerificationCode(telephone string) (string, int) {
	client, err := s.createClient()
	if err != nil {
		return "系统错误", -1
	}

	key := "auth_code_" + telephone
	code, err := Get(key)
	if err != nil {
		return "系统错误", -1
	}

	if code != "" {
		return "目前还不能发送验证码，请输入已发送的验证码", -2
	}

	// 生成验证码
	code = strconv.Itoa(GetRandomInt(6))
	fmt.Println("验证码:", code) // 开发环境打印，生产环境应该移除
	
	// 存储到 Redis
	err = Set(key, code, time.Minute)
	if err != nil {
		return "系统错误", -1
	}

	// 发送短信
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String(s.conf.SignName),
		TemplateCode:  tea.String(s.conf.TemplateCode),
		PhoneNumbers:  tea.String(telephone),
		TemplateParam: tea.String("{\"code\":\"" + code + "\"}"),
	}

	runtime := &util.RuntimeOptions{}
	rsp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		// 发送失败时删除 Redis 中的验证码
		_ = Del(key)
		return "系统错误", -1
	}

	fmt.Printf("短信发送响应: %s\n", *util.ToJSONString(rsp))
	return "验证码发送成功，请及时在对应电话查收短信", 0
}

func (s *SmsService) GetRandomInt(num int) int {
	return rand.Intn(9*int(math.Pow(10, float64(num-1)))) + int(math.Pow(10, float64(num-1)))
}

