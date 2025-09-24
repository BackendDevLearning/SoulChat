## 消息队列

### 1. 消息队列（Message Queue, MQ）

- **概念**：
   一个通用的消息传递模型，核心就是**生产者发送消息 → 消息队列存储 → 消费者接收消息**。
- **作用**：
   主要解决消息传递的问题：解耦、削峰、异步处理、消息分发等。
- **常见实现**：RabbitMQ、ActiveMQ、RocketMQ、Redis Stream、Kafka 等等。

可以理解为 **“消息队列”是一类中间件的统称**。





## Kafka

### 1. Kafka简介

- Kafka 其实就是 **消息队列的一个实现**，但更准确地说，它是一个 **分布式流式处理平台**。
- 它比传统消息队列更强：
  1. **分布式**：消息按 **分区（partition）** 存储，支持高吞吐和水平扩展。
  2. **持久化**：消息写入磁盘日志，不像有些 MQ 只存在内存。
  3. **发布订阅模型**：支持点对点（队列模式）和发布订阅（广播模式）。
  4. **流处理能力**：配合 Kafka Streams 可以直接做数据处理。

### **2. Broker 是什么？**

在 Kafka 体系中：

- **Broker** = Kafka 集群中的一个 **服务器节点**（**一个 Kafka 进程**）。
- 一个 Kafka 集群通常包含 **多个 Broker**，**每个 Broker 负责管理部分** **Topic 分区（Partition）** 的数据。

> **类比：**
>
> - Kafka 集群 = 快递公司
> - Broker = 每个分部或仓库
> - Topic = 快递产品种类
> - Partition = 具体的货架或分区
> - 消息 = 快递包裹

所以，Broker 是整个 Kafka 集群的 **核心节点**，负责：

1. 存储消息数据
2. 处理生产者写入消息
3. 处理消费者读取消息
4. 通过 **Zookeeper 或 Kafka 自带的 KRaft** 协调多个 Broker 之间的工作



### **3. 分区 Partition**

分区就是 **Topic 的水平切分**，每个 Topic 都可以包含多个分区。

#### **定义**

- **分区 = 一个有序的消息队列**。
- 分区中的消息按 **时间顺序追加写入**，消息在分区中有唯一的 **偏移量（offset）** 标识。

> Kafka 中每个分区的底层本质是 **一个可追加写入的日志文件**。

------

#### **为什么需要分区？**

##### **3.1 解决单节点性能瓶颈**

如果一个 Topic 只有 1 个分区，那么所有消息只能保存在一个 Broker 上：

- 读写压力集中在一个节点，吞吐量受限。

通过 **多个分区**：

- Kafka 可以把消息分散到多个 Broker 上，每个 Broker 同时处理一部分分区的数据。
- **并行读写**，显著提升 Kafka 集群的性能和吞吐量。

> **类比：**
>
> - 只有一个收银台 → 所有人只能排一条长队。
> - 多个收银台（分区） → 多条队伍同时处理，效率更高。

------

##### **3.2 水平扩展**

当消息量增加时，只需要增加分区数或增加 Broker 数量即可扩容。

------

#### **分区示例一**

假设有一个 Topic 叫 `orders`，包含 **3 个分区**：

| 分区            | 存储消息              |
| --------------- | --------------------- |
| **Partition 0** | Msg 1 → Msg 4 → Msg 7 |
| **Partition 1** | Msg 2 → Msg 5 → Msg 8 |
| **Partition 2** | Msg 3 → Msg 6 → Msg 9 |

Kafka 会根据 **分区策略**（例如 Key Hash、轮询 Round-Robin）将消息分配到不同分区中。

> **示意图：**

```
Topic: orders
 ├── Partition 0 → Broker 1
 ├── Partition 1 → Broker 2
 └── Partition 2 → Broker 3
```

#### **分区示例二**

orders、notifications、payments 都是不同的 topic， P 是他们的分区

```
Kafka Cluster
 ├── Broker 0
 │    ├── orders P0 (Leader)
 │    ├── notifications P0 (Leader)
 │    └── notifications P3 (Leader)
 ├── Broker 1
 │    ├── orders P1 (Leader)
 │    ├── payments P0 (Leader)
 │    └── notifications P1 (Leader)
 └── Broker 2
      ├── orders P2 (Leader)
      ├── payments P1 (Leader)
      └── notifications P2 (Leader)
```

每个 Broker 负责不同 Topic 的部分 Partition，实现负载均衡和高可用。





### **4. 副本 Replica**

副本是 **分区的备份**，用于保证 **高可用和容错**。

------

#### **定义**

- 每个分区可以配置 **多个副本**。
- 一个分区有且只有 **一个 Leader 副本**，其他为 **Follower 副本**。
- **Leader 负责读写**，**Follower 只负责同步 Leader 数据**。

------

#### **副本作用：容错和高可用**

如果某个 Broker 宕机，Kafka 可以将分区的 **Follower 副本升级为新的 Leader**，保证数据不丢失、服务不中断。

> **类比：**
>
> - 银行系统：一个主服务器（Leader）负责处理交易，备份服务器（Follower）实时同步，一旦主服务器故障，备份服务器立即接管。

------

#### **副本示例**

假设 Topic `orders` 中的 **Partition 0** 设置了 **副本因子（Replication Factor）= 3**：

| 角色          | 存储位置 |
| ------------- | -------- |
| Leader 副本   | Broker 1 |
| Follower 副本 | Broker 2 |
| Follower 副本 | Broker 3 |

**正常情况下：**

- 所有读写请求都发送给 **Broker 1（Leader）**。
- Broker 2 和 Broker 3 **实时从 Leader 同步数据**。

**Broker 1 宕机时：**

- Kafka 自动选举 Broker 2 或 Broker 3 作为新的 Leader。
- 消费者和生产者自动切换到新的 Leader。

------

### **5. 分区 + 副本 综合示例**

假设我们有一个 Topic 叫 `user_events`，配置：

- **分区数：3**
- **副本因子：2**

Kafka 会将分区和副本分配到不同 Broker：

| 分区        | Leader   | Follower |
| ----------- | -------- | -------- |
| Partition 0 | Broker 1 | Broker 2 |
| Partition 1 | Broker 2 | Broker 3 |
| Partition 2 | Broker 3 | Broker 1 |

> **拓扑图：**

```
Partition 0: [Leader → Broker 1] [Follower → Broker 2]
Partition 1: [Leader → Broker 2] [Follower → Broker 3]
Partition 2: [Leader → Broker 3] [Follower → Broker 1]
```

**特点：**

1. 数据均匀分布，负载均衡。
2. 每个分区都有至少 1 个备份，保证高可用。



### **6. ZooKeeper 在 Kafka 中的作用**

Kafka 是一个 **分布式消息系统**，其核心目标是保证高可用、高性能和分布式扩展。ZooKeeper 在其中主要扮演 **集群协调者** 的角色。

##### **核心职责**

| ZooKeeper 作用                   | 详细说明                                                     |
| -------------------------------- | ------------------------------------------------------------ |
| **1. Broker 管理（注册与发现）** | 当 Kafka Broker 启动时，会向 ZooKeeper 注册自己，存储 Broker 的 `IP` 和 `端口` 等信息。 集群中的其他组件（Producer/Consumer）通过 ZooKeeper 获取当前存活的 Broker 列表。 |
| **2. Controller 选举**           | Kafka 集群中需要有一个 **Controller Broker** 负责管理分区的 Leader 选举。 ZooKeeper 确保**只有一个 Controller 存活**，通过 **临时节点** 实现自动选举和故障转移。 |
| **3. 分区 Leader 选举**          | 当某个 Broker 宕机时，ZooKeeper 负责触发 Leader 重新选举，保证消息仍然可用。 |
| **4. Topic、分区等元数据存储**   | Topic 列表、分区信息、副本分布情况都保存在 ZooKeeper 中。 这样集群可以共享这些配置信息。 |
| **5. 消费者组管理（早期版本）**  | 在 Kafka 0.9 之前，消费者组偏移量也保存在 ZooKeeper 中。 后来迁移到 Kafka 内部的 `__consumer_offsets` 主题中。 |