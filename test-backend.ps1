Write-Host "🔍 SoulChat 后端接口测试脚本" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green

Write-Host ""
Write-Host "1. 测试后端服务状态..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8000/api/users/login" -Method POST -ContentType "application/json" -Body '{"phone":"test","password":"test"}' -ErrorAction Stop
    Write-Host "✅ 后端服务运行正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 后端服务无法访问" -ForegroundColor Red
    Write-Host "请检查后端服务是否在 8000 端口运行" -ForegroundColor Red
    Read-Host "按任意键退出"
    exit 1
}

Write-Host ""
Write-Host "2. 测试注册接口..." -ForegroundColor Yellow
Write-Host "发送注册请求..."
try {
    $registerBody = @{
        username = "testuser"
        phone = "13800138000"
        password = "123456"
    } | ConvertTo-Json
    
    $registerResponse = Invoke-WebRequest -Uri "http://localhost:8000/api/users" -Method POST -ContentType "application/json" -Body $registerBody -ErrorAction Stop
    Write-Host "✅ 注册接口测试成功" -ForegroundColor Green
    Write-Host "响应内容: $($registerResponse.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ 注册接口测试失败" -ForegroundColor Red
    Write-Host "HTTP状态码: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $responseBody = $reader.ReadToEnd()
        Write-Host "错误响应: $responseBody" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "3. 测试登录接口..." -ForegroundColor Yellow
Write-Host "发送登录请求..."
try {
    $loginBody = @{
        phone = "13800138000"
        password = "123456"
    } | ConvertTo-Json
    
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8000/api/users/login" -Method POST -ContentType "application/json" -Body $loginBody -ErrorAction Stop
    Write-Host "✅ 登录接口测试成功" -ForegroundColor Green
    Write-Host "响应内容: $($loginResponse.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ 登录接口测试失败" -ForegroundColor Red
    Write-Host "HTTP状态码: $($_.Exception.Response.StatusCode.value__)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $responseBody = $reader.ReadToEnd()
        Write-Host "错误响应: $responseBody" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "📋 测试完成！" -ForegroundColor Green
Write-Host "如果注册接口返回 500 错误，请检查：" -ForegroundColor Yellow
Write-Host "1. 数据库连接是否正常" -ForegroundColor White
Write-Host "2. Redis 连接是否正常" -ForegroundColor White
Write-Host "3. 表结构是否正确创建" -ForegroundColor White
Write-Host "4. 后端服务日志中的详细错误信息" -ForegroundColor White

Read-Host "按任意键退出"
