# SoulChat Frontend 快速启动指南

## 🎉 Vue 前端项目创建完成！

### 📁 项目结构
```
frontend/
├── src/
│   ├── api/              # API 接口层
│   ├── assets/           # 静态资源
│   ├── components/       # 可复用组件
│   ├── stores/           # Pinia 状态管理
│   ├── utils/            # 工具函数
│   ├── views/            # 页面组件
│   ├── router/           # 路由配置
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── package.json          # 项目配置
├── vite.config.js        # Vite 配置
└── README.md             # 详细文档
```

## 🚀 快速开始

### 1. 安装依赖
```bash
cd frontend
npm install
```

### 2. 启动开发服务器
```bash
npm run dev
```

访问: http://localhost:3000

### 3. 构建生产版本
```bash
npm run build
```

## 🔧 核心功能

### ✅ 已完成功能

1. **用户认证系统**
   - 登录/注册页面
   - JWT token 管理
   - 路由权限控制

2. **聊天功能**
   - WebSocket 实时通信
   - 私聊和群聊支持
   - 消息历史记录
   - 在线状态显示

3. **用户管理**
   - 个人资料编辑
   - 头像和封面图片
   - 密码修改
   - 关注系统

4. **群组管理**
   - 创建群组
   - 加入群组
   - 群组消息

5. **技术集成**
   - Protobuf 消息序列化
   - Element Plus UI 组件
   - Pinia 状态管理
   - Axios HTTP 客户端

## 🌐 API 接口

### 认证接口
- `POST /api/users/login` - 用户登录
- `POST /api/users` - 用户注册
- `GET /api/profiles/me` - 获取个人信息
- `PUT /api/users/updateUserInfo` - 更新用户信息

### 聊天接口
- `POST /api/chat/send` - 发送消息
- `GET /api/chat/messages` - 获取消息历史
- `POST /api/chat/groups` - 创建群组
- `POST /api/chat/groups/{id}/join` - 加入群组

### WebSocket 连接
- `ws://localhost:8000/ws?user={username}` - WebSocket 连接

## 🎨 页面组件

### 1. 登录页面 (`/login`)
- 手机号 + 密码登录
- 表单验证
- 自动跳转到聊天页面

### 2. 注册页面 (`/register`)
- 用户名 + 手机号 + 密码注册
- 密码确认验证
- 注册成功后自动登录

### 3. 聊天页面 (`/chat`)
- 侧边栏：聊天列表、群组列表
- 主区域：消息列表、消息输入
- 实时消息推送
- 在线用户显示

### 4. 个人资料页面 (`/profile`)
- 头像上传
- 个人信息编辑
- 密码修改
- 封面图片设置

### 5. 用户资料页面 (`/profile/:userId`)
- 查看其他用户信息
- 关注/取消关注
- 发起私聊

## 🔌 WebSocket 集成

### 消息格式
```javascript
// 发送消息
{
  to: "user123",           // 接收者
  content: "Hello!",       // 消息内容
  type: 1,                 // 1=私聊, 2=群聊
  contentType: 1,          // 1=文字, 2=图片, 3=文件
  url: ""                  // 文件URL
}

// 心跳消息
{
  type: 0,                 // HEAT_BEAT
  contentType: 1,
  content: "ping"
}
```

### 事件监听
```javascript
// 连接成功
websocketService.on('connected', () => {
  console.log('WebSocket 连接成功')
})

// 收到消息
websocketService.on('message', (message) => {
  console.log('收到消息:', message)
})

// 用户上线
websocketService.on('userJoined', (user) => {
  console.log('用户上线:', user)
})
```

## 🛠️ 开发工具

### 代码格式化
```bash
npm run lint      # ESLint 检查
npm run format    # Prettier 格式化
```

### 开发调试
- Vue DevTools 浏览器扩展
- Element Plus 组件调试
- WebSocket 连接状态监控

## 📱 移动端适配

- 响应式设计，支持移动端访问
- 触摸友好的交互设计
- 移动端优化的聊天界面

## 🔒 安全特性

- JWT token 自动刷新
- XSS 防护
- CSRF 防护
- 输入内容过滤

## 🚀 部署建议

### 开发环境
```bash
# 启动后端服务
cd ../
kratos run

# 启动前端服务
cd frontend
npm run dev
```

### 生产环境
```bash
# 构建前端
npm run build

# 使用 nginx 部署
# 配置反向代理到后端服务
```

## 📞 技术支持

如果遇到问题，请检查：

1. **后端服务是否启动** - 确保 Kratos 服务在 8000 端口运行
2. **WebSocket 连接** - 检查浏览器控制台是否有连接错误
3. **API 接口** - 确保后端 API 接口正常响应
4. **CORS 配置** - 确保后端允许前端域名访问

## 🎯 下一步开发

可以考虑添加的功能：

1. **文件上传** - 图片、文件上传功能
2. **语音消息** - 语音录制和播放
3. **视频通话** - WebRTC 视频通话
4. **消息搜索** - 消息内容搜索
5. **表情包** - 表情和贴纸支持
6. **消息撤回** - 消息撤回功能
7. **消息转发** - 消息转发功能
8. **主题切换** - 明暗主题切换

---

🎉 **恭喜！你的 Vue 前端项目已经创建完成，可以开始开发了！**
