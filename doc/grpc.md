# grpc

## 传统 rpc 的方案：

1. json + http 1.x
2. tcp + 自定义协议

### 局限性

| 方法            | 局限                                                         |
| --------------- | ------------------------------------------------------------ |
| HTTP/1.x + JSON | **单连接顺序**调用，频繁建立连接开销大；**JSON 体积大**、解析慢；不易做流式通信；**多语言调用需手写 SDK** 或处理兼容性 |
| TCP 自定义协议  | **粘包**、拆包、消息边界、并发控制都需手动实现；跨语言支持困难；**缺少统一标准** |

> 总结：传统方法**性能低、开发复杂、跨语言困难、并发能力差**。



## grpc 的目标

1. 高性能，低数据量，长连接
2. 多语言
3. 消息边界清晰
4. 并发与流式通信

###  核心功能与优化

| 功能/优化                                                    | 描述                                                         |
| ------------------------------------------------------------ | :----------------------------------------------------------- |
| [HTTP/2](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=HTTP%2F2&zhida_source=entity) [多路复用](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=多路复用&zhida_source=entity) | **单条 TCP 连接上**可以同时发送多个 RPC 请求，每个请求有唯一 [Stream ID](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=Stream+ID&zhida_source=entity)，互不干扰 |
| [Protobuf](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=Protobuf&zhida_source=entity) 序列化 | **二进制**、高效、体积小、跨语言，可直接生成代码             |
| 流式 RPC                                                     | Unary（单请求单响应）、[Server Streaming](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=Server+Streaming&zhida_source=entity)、[Client Streaming](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=Client+Streaming&zhida_source=entity)、[Bidirectional Streaming](https://zhida.zhihu.com/search?content_id=262975642&content_type=Article&match_order=1&q=Bidirectional+Streaming&zhida_source=entity) |
| **消息边界**                                                 | **每个请求和响应都有 length-prefix，自动拼装 frame，避免粘包** |
| **服务注册与方法分发**                                       | 通过 Map/反射自动分发方法，无需手写 switch-case              |
| 高并发处理                                                   | **Stream** 内部通过 Stream ID 和帧拼装保证并发调用安全       |
| **代码生成**                                                 | proto 定义 → 自动生成客户端和服务端代码 → 调用像本地函数一样 |



##  gRPC 的底层逻辑

- **物理层**：TCP 长连接
- **应用层**：HTTP/2
- **逻辑通道**：Stream（每个 RPC 调用对应一个 Stream）

gRPC 利用了 HTTP/2 的多个特性来实现高效的通信，但是不能直接等同于http协议，前端识别不了

gRPC 的请求在底层确实用的是 HTTP/2 的 Frame，但它的 **内容不是普通 HTTP 文本请求**：

- Header 是特殊的二进制字段（例如 `content-type: application/grpc`）；
- Body 是经过 **Protobuf 编码的二进制流**；
- 支持 **流式数据（Stream）**，而不是传统的一问一答。

所以从浏览器角度看，这个请求完全是“陌生格式”，一种方言 http/2，浏览器不会去帮你解析。

### http 1.1

#### http 协议基于 tcp/ip，并且使用了请求 - 应答 的通信模式

1. 长连接

   任意一端没有明确提出断开连接，则保持 tcp 连接状态

2. 管道传输

   一个请求发出去不必等其回来就可以再次发送请求

3. 队头阻塞

   相比于 http 1.0



#### 缺陷

1. 只压缩了 body
2. 服务器按顺序响应，队头阻塞
3. 服务器不能主动推送
4. 明文传播
5. 身份校验
6. 信息完整性



### http 2.0

#### 优点

1. 队头压缩

2. 二进制格式

3. 并发传输

   1. 一个 **TCP** 包含多个 **stream**，**stream** 包含多个 **Message**（对应 http 1.x 中的请求响应），Frame 是 Http/2 最小单位，以二进制压缩格式存放再 HTTP/1 中的内容
   2. 不同的 http 请求（Message 针对为不同的stream ID）用独一无二的 stream ID 区分

4. 服务端推送

   双方都建立 stream，客户端为奇数号，服务器为偶数号。



http/1.1 为长连接，为什么不能实现聊天？

因为其为半双工，同一时间客户端和服务器只能有一方主动发送数据



既然grpc走的http2，也是长连接，为什么不能实现聊天呢







1. message severce

