# WebSocket和HTTP服务对比

## 1. 基本概念

| 特性     | HTTP                                            | WebSocket                                |
| -------- | ----------------------------------------------- | ---------------------------------------- |
| 协议     | 请求-响应（Request-Response）                   | 全双工（Full-Duplex）持久连接            |
| 连接方式 | 每次请求建立一次 TCP 连接（短连接）             | 建立一次 TCP 连接后保持长连接            |
| 通信模式 | 客户端主动发起，服务端被动响应                  | 客户端和服务端都可以主动发送数据         |
| 数据格式 | 文本/二进制（通常是 JSON/XML/HTML）             | 文本/二进制（可以是 JSON、protobuf 等）  |
| 适合场景 | 请求-响应场景：浏览网页、提交表单、获取接口数据 | 实时场景：聊天、游戏、金融行情、IoT 设备 |

------

## 2. 连接特性对比

| 特性         | HTTP                                      | WebSocket                            |
| ------------ | ----------------------------------------- | ------------------------------------ |
| 连接持续时间 | 短，完成一次请求就关闭                    | 长，连接保持直到客户端或服务端关闭   |
| 资源消耗     | 每次请求都需重新建立 TCP/HTTP 连接        | 一次连接可发送多条消息，减少握手开销 |
| 延迟         | 较高（每次请求都要握手 + TCP 建立）       | 较低，长连接后无需重复握手           |
| 双向通信     | 不支持服务端主动推送（除非用 SSE/长轮询） | 原生支持双向实时推送                 |

------

## 3. 性能对比

| 特性     | HTTP                               | WebSocket                              |
| -------- | ---------------------------------- | -------------------------------------- |
| 连接开销 | 高（每次请求都要 TCP + HTTP 握手） | 低（一次握手，后续复用）               |
| 数据开销 | 较大（每次请求都有 HTTP headers）  | 较小（只有一次握手，之后是原始数据流） |
| 并发能力 | 受限于短连接频繁创建和关闭         | 高并发更友好，适合实时推送场景         |

------

## 4. 应用场景对比

| 应用类型         | HTTP                    | WebSocket             |
| ---------------- | ----------------------- | --------------------- |
| 静态网页访问     | ✅                       | ❌（不适合）           |
| REST API         | ✅                       | ❌（HTTP 更标准）      |
| 聊天系统         | ❌（可用轮询，但效率低） | ✅（实时双向）         |
| 游戏             | ❌                       | ✅（低延迟实时交互）   |
| 金融行情         | ❌                       | ✅（实时推送价格变化） |
| IoT 设备数据上报 | ✅（周期性）             | ✅（实时）             |

------

## 5. 使用方式示例

### HTTP 示例（Go）：

```go
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello HTTP"))
})
http.ListenAndServe(":8080", nil)
```

### WebSocket 示例（Go + Gorilla WebSocket）：

```go
upgrader := websocket.Upgrader{}
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        conn.WriteMessage(websocket.TextMessage, msg)
    }
})
http.ListenAndServe(":8080", nil)
```

- HTTP：客户端发请求 → 服务端回应 → 连接关闭
- WebSocket：客户端/服务端可随时互发消息 → 连接保持直到关闭

------

### 💡 总结

1. **HTTP**：短连接、请求-响应、适合静态访问和标准 API
2. **WebSocket**：长连接、双向实时通信、适合聊天、游戏、金融等实时场景



## 6. 常用后端开发场景设计

**HTTP 协议是 WebSocket 的基础**

- WebSocket 连接最开始是通过 **HTTP/1.1 请求**发起的
- 客户端发送一个特殊的 `Upgrade` 请求头（`Connection: Upgrade` + `Upgrade: websocket`）

**服务器升级连接**

- `websocket.Upgrader` 就是 Go 中处理这个升级请求的工具
- `Upgrade(w, r, nil)` 会把这个 HTTP 连接 **升级成 WebSocket 长连接**
- 升级成功后，客户端和服务端就可以进行 **双向实时通信**

**区别于普通 HTTP**

- 普通 HTTP：每次请求都是短连接，服务端不能主动推送
- WebSocket：升级后连接保持，服务端可以随时向客户端发送消息

**Web 服务同时提供 HTTP 和 WebSocket**

- HTTP 路由：提供 REST API、网页访问
- `/ws` 路由：处理实时消息，升级为 WebSocket，***目前在我们的项目中，主要使用在聊天场景下***

