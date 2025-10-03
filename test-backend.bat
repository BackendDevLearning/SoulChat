@echo off
echo 🔍 SoulChat 后端接口测试脚本
echo ================================

echo.
echo 1. 测试后端服务状态...
curl -s http://localhost:8000/api/users/login -X POST -H "Content-Type: application/json" -d "{\"phone\":\"test\",\"password\":\"test\"}" > nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 后端服务运行正常
) else (
    echo ❌ 后端服务无法访问
    echo 请检查后端服务是否在 8000 端口运行
    pause
    exit /b 1
)

echo.
echo 2. 测试注册接口...
echo 发送注册请求...
curl -X POST http://localhost:8000/api/users ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"testuser\",\"phone\":\"13800138000\",\"password\":\"123456\"}" ^
  -w "\nHTTP状态码: %%{http_code}\n"

echo.
echo 3. 测试登录接口...
echo 发送登录请求...
curl -X POST http://localhost:8000/api/users/login ^
  -H "Content-Type: application/json" ^
  -d "{\"phone\":\"13800138000\",\"password\":\"123456\"}" ^
  -w "\nHTTP状态码: %%{http_code}\n"

echo.
echo 📋 测试完成！
echo 如果注册接口返回 500 错误，请检查：
echo 1. 数据库连接是否正常
echo 2. Redis 连接是否正常
echo 3. 表结构是否正确创建
echo 4. 后端服务日志中的详细错误信息

pause
