# SoulChat 后端连接测试指南

## 🎉 后端服务已成功启动！

从你的输出可以看到：
```
✅ HTTP 服务: http://localhost:8000
✅ gRPC 服务: localhost:9000  
✅ Kafka 服务: 192.168.218.131:9092
✅ WebSocket 服务: ws://localhost:8000/ws
```

## 🔧 前端测试方案

### 方案1：使用测试页面（无需 Node.js）

1. **打开测试页面**
   ```
   在浏览器中打开: frontend/test.html
   ```

2. **测试功能**
   - ✅ API 接口连接测试
   - ✅ WebSocket 连接测试
   - ✅ 消息发送测试
   - ✅ 注册/登录接口测试

### 方案2：安装 Node.js 运行完整前端

1. **下载安装 Node.js**
   - 访问: https://nodejs.org/
   - 下载 LTS 版本
   - 安装时选择 "Add to PATH"

2. **启动前端服务**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

3. **访问应用**
   ```
   http://localhost:3000
   ```

## 🌐 端口对应关系

| 服务 | 端口 | 说明 |
|------|------|------|
| 后端 HTTP API | 8000 | REST API 接口 |
| 后端 gRPC | 9000 | gRPC 服务 |
| 前端开发服务器 | 3000 | Vue 开发服务器 |
| WebSocket | 8000/ws | 实时消息通信 |
| Kafka | 9092 | 消息队列 |

## 🔌 API 接口测试

### 测试注册接口
```bash
curl -X POST http://localhost:8000/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "phone": "13800138000", 
    "password": "123456"
  }'
```

### 测试登录接口
```bash
curl -X POST http://localhost:8000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "password": "123456"
  }'
```

### 测试 WebSocket 连接
```javascript
const ws = new WebSocket('ws://localhost:8000/ws?user=testuser');
ws.onopen = () => console.log('WebSocket 连接成功');
ws.onmessage = (event) => console.log('收到消息:', event.data);
```

## 🎯 快速测试步骤

1. **打开测试页面**
   ```
   双击 frontend/test.html 文件
   ```

2. **自动测试**
   - 页面会自动测试 API 连接
   - 点击 "连接 WebSocket" 测试实时通信
   - 输入消息测试消息发送

3. **查看结果**
   - 绿色 ✅ 表示成功
   - 红色 ❌ 表示失败
   - 蓝色 ℹ️ 表示信息

## 📱 功能验证清单

- [ ] API 接口响应正常
- [ ] WebSocket 连接成功
- [ ] 消息发送接收正常
- [ ] 用户注册功能
- [ ] 用户登录功能
- [ ] Kafka 消息队列工作正常

## 🚀 下一步开发

1. **安装 Node.js** 运行完整前端项目
2. **开发新功能** 基于现有架构
3. **集成测试** 前后端联调
4. **部署上线** 生产环境配置

## 🔍 常见问题

### Q: npm 命令无法识别？
A: 需要安装 Node.js，或使用测试页面进行功能验证

### Q: WebSocket 连接失败？
A: 检查后端服务是否在 8000 端口运行，防火墙是否阻止连接

### Q: API 请求失败？
A: 检查后端服务状态，确认端口和路径正确

---

🎉 **你的后端服务已经完美运行，现在可以开始测试前端功能了！**
