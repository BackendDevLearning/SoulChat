# SoulChat Frontend

基于 Vue 3 + Element Plus 的现代化聊天应用前端，配合 Kratos 后端框架和 protobuf 协议。

## 🚀 特性

- **Vue 3** - 使用最新的 Vue 3 Composition API
- **Element Plus** - 现代化的 UI 组件库
- **Pinia** - 轻量级状态管理
- **Vue Router** - 官方路由管理器
- **WebSocket** - 实时消息通信
- **Protobuf** - 高效的数据序列化
- **Axios** - HTTP 请求库
- **Vite** - 快速的构建工具

## 📦 安装

```bash
# 安装依赖
npm install

# 或者使用 yarn
yarn install
```

## 🛠️ 开发

```bash
# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

## 🏗️ 项目结构

```
frontend/
├── src/
│   ├── api/              # API 接口
│   │   └── index.js      # API 配置和接口定义
│   ├── assets/           # 静态资源
│   ├── components/       # 可复用组件
│   ├── stores/           # Pinia 状态管理
│   │   ├── user.js       # 用户状态
│   │   └── chat.js       # 聊天状态
│   ├── utils/            # 工具函数
│   │   ├── index.js      # 通用工具
│   │   └── websocket.js  # WebSocket 服务
│   ├── views/            # 页面组件
│   │   ├── Login.vue     # 登录页
│   │   ├── Register.vue  # 注册页
│   │   ├── Chat.vue      # 聊天页
│   │   ├── Profile.vue   # 个人资料页
│   │   └── UserProfile.vue # 用户资料页
│   ├── router/           # 路由配置
│   │   └── index.js      # 路由定义
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── index.html            # HTML 模板
├── package.json          # 项目配置
├── vite.config.js        # Vite 配置
└── README.md             # 项目说明
```

## 🔧 配置

### 环境配置

项目使用 Vite 的代理功能，开发环境下的 API 请求会自动代理到后端服务器：

```javascript
// vite.config.js
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8000',  // 后端服务器地址
      changeOrigin: true,
    },
    '/ws': {
      target: 'ws://localhost:8000',    // WebSocket 服务器地址
      ws: true,
    },
  },
}
```

### API 配置

API 基础配置在 `src/api/index.js` 中：

```javascript
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})
```

## 🌐 功能模块

### 1. 用户认证

- **登录/注册** - 支持手机号登录注册
- **JWT 认证** - 基于 JWT 的身份验证
- **自动登录** - 本地存储 token 自动登录
- **权限控制** - 路由级别的权限控制

### 2. 聊天功能

- **实时消息** - WebSocket 实时消息推送
- **私聊/群聊** - 支持单聊和群聊
- **消息类型** - 文字、图片、文件、语音、视频
- **消息历史** - 消息记录和搜索
- **在线状态** - 用户在线状态显示

### 3. 用户管理

- **个人资料** - 用户信息编辑
- **头像上传** - 支持头像和封面图片
- **密码修改** - 安全的密码修改
- **关注系统** - 用户关注和粉丝

### 4. 群组管理

- **创建群组** - 创建和管理群组
- **加入群组** - 邀请和加入群组
- **群组设置** - 群组信息和权限管理

## 🔌 WebSocket 集成

### 连接管理

```javascript
// 连接 WebSocket
websocketService.connect(username)

// 监听事件
websocketService.on('message', (message) => {
  console.log('收到消息:', message)
})

// 发送消息
websocketService.sendMessage({
  to: 'user123',
  content: 'Hello!',
  type: 1,
  contentType: 1
})
```

### Protobuf 支持

项目集成了 protobuf 支持，可以高效地序列化和反序列化消息：

```javascript
// 发送 protobuf 消息
const message = messageType.create(messageData)
const buffer = messageType.encode(message).finish()
websocket.send(buffer)

// 接收 protobuf 消息
const message = messageType.decode(new Uint8Array(data))
const messageObj = messageType.toObject(message)
```

## 🎨 UI 组件

### Element Plus 集成

项目使用 Element Plus 作为 UI 组件库，支持自动导入：

```javascript
// vite.config.js
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
})
```

### 响应式设计

- **移动端适配** - 支持移动端和桌面端
- **响应式布局** - 自适应不同屏幕尺寸
- **暗色主题** - 支持明暗主题切换

## 📱 移动端支持

项目支持移动端访问，主要特性：

- **触摸友好** - 优化的触摸交互
- **手势支持** - 支持滑动等手势操作
- **移动端适配** - 响应式布局适配移动设备

## 🔒 安全特性

- **XSS 防护** - 输入内容过滤和转义
- **CSRF 防护** - 请求头验证
- **内容安全策略** - CSP 头部配置
- **HTTPS 支持** - 生产环境 HTTPS

## 🚀 部署

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `dist` 目录。

### 部署到服务器

```bash
# 使用 nginx 部署
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8000;
    }
    
    location /ws {
        proxy_pass http://localhost:8000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## 🧪 测试

```bash
# 运行测试
npm run test

# 代码检查
npm run lint

# 格式化代码
npm run format
```

## 📝 开发规范

### 代码风格

- 使用 ESLint + Prettier 进行代码格式化
- 遵循 Vue 3 Composition API 最佳实践
- 使用 TypeScript 进行类型检查（可选）

### 提交规范

```bash
# 功能开发
git commit -m "feat: 添加用户登录功能"

# 问题修复
git commit -m "fix: 修复消息发送失败问题"

# 文档更新
git commit -m "docs: 更新 API 文档"
```

## 🤝 贡献

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 联系方式

- 项目链接: [https://github.com/your-username/soulchat-frontend](https://github.com/your-username/soulchat-frontend)
- 问题反馈: [Issues](https://github.com/your-username/soulchat-frontend/issues)

## 🙏 致谢

- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - Vue 3 UI 组件库
- [Pinia](https://pinia.vuejs.org/) - Vue 状态管理库
- [Vite](https://vitejs.dev/) - 下一代前端构建工具
