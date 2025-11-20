# Kafka 配置说明

## 配置项说明

### 基础配置
- `enabled`: 是否启用 Kafka（bool）
- `hosts`: Kafka broker 地址，多个地址用逗号分隔（string）
  - 示例：`"192.168.1.1:9092,192.168.1.2:9092"`
- `topic`: 主题名称（string）

### 性能配置
- `timeout`: 超时时间（秒），默认 10（int32）
- `partition`: 分区数，默认 1（int32）
- `replication_factor`: 副本因子，默认 1（int32）

### 消费者配置
- `group_id`: 消费者组 ID，默认 "chat"（string）
- `start_offset`: 起始消费位置（string）
  - `"first"`: 从最早的消息开始消费
  - `"last"`: 从最新的消息开始消费（默认）
- `commit_interval`: 提交偏移量的时间间隔（秒），默认与 timeout 相同（int32）

### 生产者配置
- `required_acks`: 消息确认机制（string）
  - `"none"`: 不需要确认（默认，性能最高，但可能丢失消息）
  - `"one"`: 只需要 leader 确认（平衡性能和可靠性）
  - `"all"`: 需要所有副本确认（最可靠，但性能较低）

### 主题配置
- `allow_auto_topic_creation`: 是否允许自动创建主题，默认 false（bool）

## 使用示例

### config.yaml 配置示例

```yaml
data:
  kafka:
    enabled: true
    hosts: "192.168.218.131:9092"
    topic: "go-chat-message"
    timeout: 10
    partition: 3
    group_id: "chat-consumer"
    start_offset: "last"
    replication_factor: 2
    allow_auto_topic_creation: false
    required_acks: "one"
    commit_interval: 5
```

## 注意事项

1. **重新生成 protobuf 文件**：修改 `conf.proto` 后需要运行：
   ```bash
   make config
   # 或
   protoc --proto_path=. --go_out=paths=source_relative:. internal/conf/conf.proto
   ```

2. **性能调优建议**：
   - 高吞吐量场景：`required_acks: "none"`，增加 `partition` 数量
   - 高可靠性场景：`required_acks: "all"`，增加 `replication_factor`
   - 实时性要求高：减小 `commit_interval`

3. **消费者组**：
   - 同一 `group_id` 的多个消费者会共享消费进度
   - 不同 `group_id` 的消费者会独立消费所有消息

4. **起始偏移量**：
   - `first`：适合需要处理历史消息的场景
   - `last`：适合只处理新消息的场景（默认）

