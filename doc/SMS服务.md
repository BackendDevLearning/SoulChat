# SMS短信服务

## 1. 服务介绍

**SMS**（Short Message Service，短消息服务）是指通过专业的云服务平台向用户手机发送短信的服务。本项目使用 [**阿里云的SMS SDK**](https://help.aliyun.com/zh/sms/developer-reference/sdk-product-overview/?spm=a2c4g.11186623.help-menu-44282.d_5_3.4dcc64499dPDGE&scm=20140722.H_215758._.OR_help-T_cn~zh-V_1) 实现。

## 2. 服务实现

### 2.1 配置

- protobuf定义：`internal/conf/conf.proto` → `message Sms`

- 自定义配置项：`configs/config.yaml`

```go
message Sms {
  string access_key_id = 1;    // 阿里云访问密钥
  string access_key_secret = 2;
  string endpoint = 3;         // 短信服务结点, 区分国内和国际
  string sign_name = 4;        // 短信签名
  string template_code = 5;    // 短信模板CODE

  // 验证码配置
  message VerificationCode {
    int32 length = 1;          // 验证码长度, 默认6
    string expire = 2;         // 过期时间, 如 "5m", "10m"
    int32 max_retry = 3;       // 用户最大重试次数
    bool debug = 4;            // 调试模式, 打印验证码到日志
  }

  VerificationCode verification_code = 6;

  // 限流配置
  message RateLimit {
    int32 max_requests_per_minute = 1;  // 每分钟最大请求数
    int32 max_requests_per_hour = 2;    // 每小时最大请求数
    int32 max_requests_per_day = 3;     // 每天最大请求数
  }

  RateLimit rate_limit = 7;

  // 重试配置
  message Retry {
    int32 max_attempts = 1;     // 系统最大重试次数
    string backoff = 2;         // 退避策略, 如 "1s", "2s"
  }

  Retry retry = 8;
}
```

### 2.2 初始化

SMS服务结构体

```go
type SmsService struct {
	conf       *conf.Sms            // 初始短信配置
	smsClient  *dysmsapi.Client     // 创建的阿里云短信客户端
	clientOnce sync.Once            // 确保客户端只创建一次
	clientErr  error                // 错误缓存
}

func NewSmsService(smsConf *conf.Sms) *SmsService {
	return &SmsService{
		conf: smsConf,
	}
}
```

这里使用 **懒加载** ，只在第一次发送短信使用的时候才创建

```go
// createClient 创建短信客户端
func (s *SmsService) createClient() (*dysmsapi.Client, error) {
	s.clientOnce.Do(func() {
		config := &openapi.Config{
			AccessKeyId:     tea.String(s.conf.AccessKeyId),
			AccessKeySecret: tea.String(s.conf.AccessKeySecret),
			Endpoint:        tea.String(s.conf.Endpoint),
		}
		s.smsClient, s.clientErr = dysmsapi.NewClient(config)
	})
	return s.smsClient, s.clientErr
}
```

### 2.3 RESTful API定义

为了保证业务逻辑清晰，使用 **短信验证码登录** 和 **密码登录** 这两个接口相互独立，同时也方便后续扩展登录方式，比如通过微信、QQ等 **第三方平台登录** 。

```go
  rpc Login(LoginRequest) returns (LoginReply) {
    option (google.api.http) = {
      post : "/api/users/login",
      body : "*",
    };
  }

  rpc LoginBySms(LoginBySmsRequest) returns (LoginReply) {
    option (google.api.http) = {
      post : "/api/users/login/sms",
      body : "*",
    };
  }

  rpc SendSms(SendSmsRequest) returns (SendSmsReply) {
    option (google.api.http) = {
      post : "/api/users/sendSms",
      body : "*",
    };
  }
```

### 2.4 SMS服务封装

在proto中定义协议后，分别在service层和biz层实现对应接口，在data层中 **实现SMS中间件的依赖注入** 和 **验证码缓存** 。SMS服务封装到一个独立repo，这样写看起来层数比较复杂，但是后续如果考虑将阿里云更换为其它平台SDK，就只需要修改中间件，并且实现中间件验证码发送和底层data模型redis缓存写入分离。

```go
// Biz层
type SmsRepo interface {
	SendCode(ctx context.Context, phone string) (string, error)
	SaveCode(ctx context.Context, phone, code string) error
	VerifyCode(ctx context.Context, phone, inputCode string) (bool, error)
}

// Data层
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

// 依赖注入
var ProviderSet = wire.NewSet(
	NewSmsRepo,
	sms.NewSmsService,
)
```

### 2.5 中间件实现

```go
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
```

## 3. 业务场景

### 3.1 验证码短信

- **用户注册/登录时的身份验证**
- 支付确认
- **密码重置**
- 安全操作确认，如更换绑定手机号码，注销账号

### 3.2 **通知提醒短信**

- 订单状态通知
- 物流信息提醒
- 系统告警通知
- 账户变动提醒

### 3.3 **营销推广短信**

- 促销活动通知
- 新品上市推广
- 会员关怀消息

