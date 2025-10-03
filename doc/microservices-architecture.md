# SoulChat 微服务架构设计

## 🏗️ 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                          │
│                    (Kong/Nginx/Envoy)                       │
└─────────────────┬─────────────────┬─────────────────┬───────┘
                  │                 │                 │
    ┌─────────────▼─────────────┐   │   ┌─────────────▼─────────────┐
    │      User Service        │   │   │     Profile Service       │
    │  ┌─────────────────────┐  │   │   │  ┌─────────────────────┐  │
    │  │   Authentication   │  │   │   │  │   Profile Mgmt     │  │
    │  │   User Management  │  │   │   │  │   Follow/Unfollow  │  │
    │  │   JWT Token       │  │   │   │  │   User Relations   │  │
    │  └─────────────────────┘  │   │   │  └─────────────────────┘  │
    │  ┌─────────────────────┐  │   │   │  ┌─────────────────────┐  │
    │  │   User Database    │  │   │   │  │  Profile Database   │  │
    │  └─────────────────────┘  │   │   │  └─────────────────────┘  │
    └─────────────────────────────┘   │   └─────────────────────────────┘
                                     │
    ┌─────────────▼─────────────┐   │   ┌─────────────▼─────────────┐
    │      Chat Service        │   │   │   Notification Service     │
    │  ┌─────────────────────┐  │   │   │  ┌─────────────────────┐  │
    │  │   WebSocket Hub     │  │   │   │  │   Push Notifications │  │
    │  │   Message Routing  │  │   │   │  │   Email/SMS        │  │
    │  │   Group Management  │  │   │   │  │   System Alerts     │  │
    │  └─────────────────────┘  │   │   │  └─────────────────────┘  │
    │  ┌─────────────────────┐  │   │   │  ┌─────────────────────┐  │
    │  │   Chat Database    │  │   │   │  │ Notification DB     │  │
    │  └─────────────────────┘  │   │   │  └─────────────────────┘  │
    └─────────────────────────────┘   │   └─────────────────────────────┘
                                     │
    ┌─────────────────────────────────▼─────────────────────────────────┐
    │                    Shared Infrastructure                         │
    │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐  │
    │  │   Redis     │ │   Kafka     │ │   MySQL     │ │   MongoDB   │  │
    │  │  (Cache)    │ │ (Messaging) │ │ (Primary)   │ │ (Logs)      │  │
    │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘  │
    └─────────────────────────────────────────────────────────────────────┘
```

## 📁 项目结构

```
soulchat-microservices/
├── api-gateway/                    # API 网关
│   ├── kong/
│   │   ├── kong.yml
│   │   └── plugins/
│   └── nginx/
│       └── nginx.conf
├── services/
│   ├── user-service/              # 用户服务
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── biz/
│   │   │   ├── data/
│   │   │   ├── service/
│   │   │   └── server/
│   │   ├── api/
│   │   │   └── user/
│   │   │       └── v1/
│   │   ├── configs/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── profile-service/           # 个人资料服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── api/
│   │   ├── configs/
│   │   └── Dockerfile
│   ├── chat-service/              # 聊天服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── api/
│   │   ├── configs/
│   │   └── Dockerfile
│   └── notification-service/      # 通知服务
│       ├── cmd/
│       ├── internal/
│       ├── api/
│       ├── configs/
│       └── Dockerfile
├── shared/                        # 共享组件
│   ├── proto/                     # 共享 protobuf 定义
│   ├── middleware/                # 共享中间件
│   ├── utils/                     # 共享工具
│   └── config/                    # 共享配置
├── infrastructure/               # 基础设施
│   ├── docker-compose.yml
│   ├── k8s/                      # Kubernetes 配置
│   ├── monitoring/               # 监控配置
│   └── logging/                  # 日志配置
└── docs/                         # 文档
    ├── api/
    ├── architecture/
    └── deployment/
```

## 🔄 服务间通信

### 1. 同步通信 (gRPC)
```go
// 用户服务调用个人资料服务
type ProfileClient struct {
    conn *grpc.ClientConn
}

func (c *ProfileClient) GetProfile(ctx context.Context, username string) (*Profile, error) {
    client := profilepb.NewProfileServiceClient(c.conn)
    resp, err := client.GetProfile(ctx, &profilepb.GetProfileRequest{
        Username: username,
    })
    return resp.Profile, err
}
```

### 2. 异步通信 (Kafka)
```go
// 聊天服务发送消息事件
type MessageEvent struct {
    UserID    string `json:"user_id"`
    MessageID string `json:"message_id"`
    Content   string `json:"content"`
    Timestamp int64  `json:"timestamp"`
}

func (s *ChatService) SendMessage(ctx context.Context, msg *Message) error {
    // 保存消息到数据库
    err := s.messageRepo.SaveMessage(msg)
    if err != nil {
        return err
    }
    
    // 发送事件到 Kafka
    event := &MessageEvent{
        UserID:    msg.UserID,
        MessageID: msg.ID,
        Content:   msg.Content,
        Timestamp: time.Now().Unix(),
    }
    
    return s.kafkaProducer.SendMessage("message.created", event)
}
```

## 🗄️ 数据拆分策略

### 1. 数据库拆分
```sql
-- User Service Database
CREATE DATABASE user_service;
USE user_service;
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Profile Service Database  
CREATE DATABASE profile_service;
USE profile_service;
CREATE TABLE profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    bio TEXT,
    image_url VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE follow_relations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    follower_id BIGINT NOT NULL,
    following_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_follow (follower_id, following_id)
);

-- Chat Service Database
CREATE DATABASE chat_service;
USE chat_service;
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    sender_id BIGINT NOT NULL,
    receiver_id BIGINT,
    group_id BIGINT,
    content TEXT NOT NULL,
    message_type ENUM('text', 'image', 'file') DEFAULT 'text',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. 数据一致性策略

#### Saga 模式
```go
type CreateUserSaga struct {
    userService    *UserService
    profileService *ProfileService
    notificationService *NotificationService
}

func (s *CreateUserSaga) Execute(ctx context.Context, req *CreateUserRequest) error {
    // Step 1: Create user
    user, err := s.userService.CreateUser(ctx, req)
    if err != nil {
        return err
    }
    
    // Step 2: Create profile
    profile, err := s.profileService.CreateProfile(ctx, user.ID)
    if err != nil {
        // Compensate: Delete user
        s.userService.DeleteUser(ctx, user.ID)
        return err
    }
    
    // Step 3: Send welcome notification
    err = s.notificationService.SendWelcomeNotification(ctx, user.ID)
    if err != nil {
        // Log error but don't fail the transaction
        log.Errorf("Failed to send welcome notification: %v", err)
    }
    
    return nil
}
```

## 🚀 部署策略

### 1. Docker Compose (开发环境)
```yaml
version: '3.8'
services:
  # API Gateway
  api-gateway:
    image: kong:latest
    ports:
      - "8000:8000"
      - "8001:8001"
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: /kong/kong.yml
    volumes:
      - ./api-gateway/kong/kong.yml:/kong/kong.yml
    
  # User Service
  user-service:
    build: ./services/user-service
    ports:
      - "8001:8001"
    environment:
      - DATABASE_URL=mysql://root:123456@mysql:3306/user_service
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - mysql
      - redis
      
  # Profile Service  
  profile-service:
    build: ./services/profile-service
    ports:
      - "8002:8002"
    environment:
      - DATABASE_URL=mysql://root:123456@mysql:3306/profile_service
      - REDIS_URL=redis://redis:6379/1
    depends_on:
      - mysql
      - redis
      
  # Chat Service
  chat-service:
    build: ./services/chat-service
    ports:
      - "8003:8003"
    environment:
      - DATABASE_URL=mysql://root:123456@mysql:3306/chat_service
      - REDIS_URL=redis://redis:6379/2
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - mysql
      - redis
      - kafka
      
  # Infrastructure
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: 123456
    volumes:
      - mysql_data:/var/lib/mysql
      
  redis:
    image: redis:6.2
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    
  kafka:
    image: wurstmeister/kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    depends_on:
      - zookeeper
      
  zookeeper:
    image: zookeeper
    ports:
      - "2181:2181"

volumes:
  mysql_data:
  redis_data:
```

### 2. Kubernetes (生产环境)
```yaml
# user-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: soulchat/user-service:latest
        ports:
        - containerPort: 8001
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: user-service-secret
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: user-service-secret
              key: redis-url
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 8001
    targetPort: 8001
```

## 📊 监控与日志

### 1. 服务监控
```go
// Prometheus 指标
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)
```

### 2. 分布式追踪
```go
// Jaeger 追踪
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "CreateUser")
    defer span.Finish()
    
    span.SetTag("user.username", req.Username)
    span.SetTag("user.email", req.Email)
    
    // 业务逻辑
    user, err := s.userRepo.CreateUser(ctx, req)
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(log.Error(err))
        return nil, err
    }
    
    span.SetTag("user.id", user.ID)
    return user, nil
}
```

## 🔧 迁移步骤

### 阶段 1: 准备阶段
1. **代码重构**: 将业务逻辑按服务边界重新组织
2. **API 设计**: 定义服务间接口
3. **数据模型**: 设计独立的数据模型

### 阶段 2: 服务拆分
1. **用户服务**: 先拆分用户认证和用户管理
2. **个人资料服务**: 拆分用户资料和关注功能
3. **聊天服务**: 拆分消息和群组功能
4. **通知服务**: 拆分通知和推送功能

### 阶段 3: 基础设施
1. **API 网关**: 部署 Kong 或 Nginx
2. **服务发现**: 使用 Consul 或 Kubernetes Service
3. **配置管理**: 使用 Consul 或 Kubernetes ConfigMap
4. **监控系统**: 部署 Prometheus + Grafana

### 阶段 4: 优化
1. **性能优化**: 缓存策略、数据库优化
2. **安全加固**: 服务间认证、API 限流
3. **容错处理**: 熔断器、重试机制
4. **自动化**: CI/CD 流水线

## 💡 最佳实践

### 1. 服务设计原则
- **单一职责**: 每个服务只负责一个业务域
- **数据独立**: 每个服务拥有独立的数据存储
- **接口稳定**: 服务间接口要保持向后兼容
- **故障隔离**: 服务故障不应影响其他服务

### 2. 通信模式
- **同步通信**: 用于实时性要求高的场景
- **异步通信**: 用于解耦和提升性能
- **事件驱动**: 用于松耦合的业务流程

### 3. 数据一致性
- **最终一致性**: 接受短暂的数据不一致
- **Saga 模式**: 处理跨服务事务
- **事件溯源**: 记录所有状态变更

这个微服务架构设计为你的 SoulChat 项目提供了完整的拆分方案，可以根据业务需求逐步实施。
