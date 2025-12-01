# Redis 代码解析：`myredis.GetKeyNilIsErr`

## 代码位置
```go
rspString, err = myredis.GetKeyNilIsErr("message_list_" + message.SendId + "_" + message.ReceiveId)
```

## 代码解析

### 1. **方法名称：`GetKeyNilIsErr`**

这是一个自定义的 Redis 操作方法，从命名可以推断其功能：
- `GetKey`: 获取 Redis 中的 key 值
- `NilIsErr`: 将 `redis.Nil`（key 不存在）当作错误返回

**预期行为：**
- 如果 key **存在**：返回 value 和 `err == nil`
- 如果 key **不存在**：返回空字符串和 `err == redis.Nil`
- 如果发生**其他错误**：返回错误信息

### 2. **Redis Key 的构造**

```go
"message_list_" + message.SendId + "_" + message.ReceiveId
```

**Key 格式：** `message_list_{发送者ID}_{接收者ID}`

**示例：**
- 用户 `U123` 发送给用户 `U456` 的消息列表
- Key: `message_list_U123_U456`

**设计目的：**
- 为每对用户（发送者-接收者）维护一个独立的消息列表缓存
- 支持双向对话：`message_list_U123_U456` 和 `message_list_U456_U123` 是两个不同的 key

### 3. **完整代码流程**

```go
// 1. 尝试从 Redis 获取消息列表
var rspString string
rspString, err = myredis.GetKeyNilIsErr("message_list_" + message.SendId + "_" + message.ReceiveId)

// 2. 如果 key 存在（err == nil），说明缓存中有数据
if err == nil {
    // 2.1 反序列化 JSON 字符串为消息列表
    var rsp []GetMessageListRespond
    if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
        zlog.Error(err.Error())
    }
    
    // 2.2 将新消息追加到列表
    rsp = append(rsp, messageRsp)
    
    // 2.3 重新序列化为 JSON
    rspByte, err := json.Marshal(rsp)
    if err != nil {
        zlog.Error(err.Error())
    }
    
    // 2.4 更新 Redis 缓存（带过期时间）
    if err := myredis.SetKeyEx(
        "message_list_"+message.SendId+"_"+message.ReceiveId, 
        string(rspByte), 
        time.Minute*constants.REDIS_TIMEOUT
    ); err != nil {
        zlog.Error(err.Error())
    }
} else {
    // 3. 如果 key 不存在（err == redis.Nil），不处理缓存
    //    这是正常的，说明这是第一次缓存，或者缓存已过期
    if !errors.Is(err, redis.Nil) {
        // 只有非 redis.Nil 的错误才记录日志
        zlog.Error(err.Error())
    }
}
```

## 业务逻辑说明

### 缓存策略：**写时更新（Write-Through）**

1. **新消息到达时：**
   - 先通过 WebSocket 实时推送给在线用户
   - 同时更新 Redis 缓存中的消息列表

2. **缓存的作用：**
   - **加速查询**：用户切换聊天对象时，从 Redis 快速获取消息列表
   - **减少数据库压力**：避免频繁查询数据库
   - **临时存储**：缓存最近的消息列表（有过期时间）

3. **为什么 key 不存在时不处理？**
   - 第一次发送消息时，Redis 中没有缓存是正常的
   - 缓存可能已过期，这是预期行为
   - 下次查询时会从数据库加载并重建缓存

## 等价的实现方式

### 使用当前项目的 Redis 客户端

```go
// 假设使用 model.Data.Cache() 获取 Redis 客户端
cache := data.Cache()
ctx := context.Background()

// 1. 获取消息列表
rspString, exists, err := cache.Get(ctx, "message_list_"+message.SendId+"_"+message.ReceiveId)
if err != nil {
    // 处理错误（非 redis.Nil）
    logger.Error("failed to get from redis", zap.Error(err))
} else if exists {
    // 2. key 存在，更新缓存
    var rsp []GetMessageListRespond
    if err := json.Unmarshal([]byte(rspString), &rsp); err != nil {
        logger.Error("failed to unmarshal", zap.Error(err))
    } else {
        rsp = append(rsp, messageRsp)
        rspByte, _ := json.Marshal(rsp)
        
        // 更新缓存，设置过期时间
        cache.Set(ctx, 
            "message_list_"+message.SendId+"_"+message.ReceiveId, 
            string(rspByte), 
            time.Minute*REDIS_TIMEOUT,
        )
    }
}
// 3. key 不存在时，不处理（正常情况）
```

## 注意事项

1. **数据一致性：**
   - Redis 缓存是临时数据，数据库是持久化存储
   - 缓存过期或丢失时，会从数据库重新加载

2. **Key 命名规范：**
   - 使用前缀 `message_list_` 避免与其他 key 冲突
   - 使用 `SendId_ReceiveId` 标识对话双方

3. **性能考虑：**
   - 消息列表可能很大，需要考虑序列化/反序列化开销
   - 可以限制缓存的消息数量（如最近 100 条）

4. **错误处理：**
   - `redis.Nil` 是正常情况（key 不存在），不需要特殊处理
   - 其他错误（网络、连接等）需要记录日志

## 改进建议

1. **使用 Redis List 而不是 String：**
   ```go
   // 使用 LPUSH/RPUSH 追加消息，LRANGE 获取列表
   cache.LPush(ctx, key, messageJson)
   cache.LTrim(ctx, key, 0, 99) // 只保留最近 100 条
   ```

2. **批量操作：**
   ```go
   // 使用 Pipeline 减少网络往返
   pipe := cache.Pipeline(ctx, func(p redis.Pipeliner) error {
       p.Get(ctx, key)
       p.Set(ctx, key, value, ttl)
       return nil
   })
   ```

3. **添加缓存预热：**
   ```go
   // 用户登录时，预加载常用对话的消息列表
   ```



