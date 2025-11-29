package kafka

import (
	"context"
	"fmt"
	"kratos-realworld/internal/conf"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
)

var ctx = context.Background()

type kafkaService struct {
	ChatWriter *kafka.Writer
	ChatReader *kafka.Reader
	KafkaConn  *kafka.Conn
	logger     *log.Helper
}

var KafkaService = new(kafkaService)

// KafkaInit 初始化kafka
func (k *kafkaService) KafkaInit(kafkaConfig *conf.Data_Kafka, logger *log.Helper) error {
	k.logger = logger
	if kafkaConfig == nil {
		return fmt.Errorf("kafka config is nil")
	}

	// 解析 hosts，支持逗号分隔的多个地址
	hosts := strings.Split(kafkaConfig.Hosts, ",")
	if len(hosts) == 0 {
		return fmt.Errorf("kafka hosts is empty")
	}

	// 从配置读取超时时间，默认 10 秒
	timeout := int(kafkaConfig.Timeout)
	if timeout <= 0 {
		timeout = 10
	}

	// 从配置读取消费者组 ID，默认 "chat"
	groupID := kafkaConfig.GroupId
	if groupID == "" {
		groupID = "chat"
	}

	// 从配置读取提交间隔，默认与 timeout 相同
	commitInterval := int(kafkaConfig.CommitInterval)
	if commitInterval <= 0 {
		commitInterval = timeout
	}

	// 解析起始偏移量
	startOffset := kafka.LastOffset
	if kafkaConfig.StartOffset == "first" {
		startOffset = kafka.FirstOffset
	}

	// 解析消息确认机制
	var requiredAcks kafka.RequiredAcks
	switch kafkaConfig.RequiredAcks {
	case "one":
		requiredAcks = kafka.RequireOne
	case "all":
		requiredAcks = kafka.RequireAll
	default:
		requiredAcks = kafka.RequireNone
	}

	// 创建 Writer
	k.ChatWriter = &kafka.Writer{
		Addr:  kafka.TCP(hosts...),
		Topic: kafkaConfig.Topic,
		// 分区策略，负载均衡
		Balancer: &kafka.Hash{},
		// 写入超时时间
		WriteTimeout: time.Duration(timeout) * time.Second,
		// 消息确认机制
		RequiredAcks: requiredAcks,
		// 是否允许自动创建主题
		AllowAutoTopicCreation: kafkaConfig.AllowAutoTopicCreation,
	}

	// 创建 Reader
	k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: hosts,
		Topic:   kafkaConfig.Topic,
		// 提交偏移量（Offset）的时间间隔，相当于提交消费的进度
		CommitInterval: time.Duration(commitInterval) * time.Second,
		// 同一 GroupID 的消费者共享消费进度，相当于共享消费的进度
		// 用于负载均衡和故障恢复，当一个消费者宕机时，其他消费者可以继续消费
		GroupID: groupID,
		// 首次启动时的起始消费位置
		StartOffset: startOffset,
	})

	if logger != nil {
		logger.Log(log.LevelInfo,
			"msg", "kafka service initialized",
			"hosts", kafkaConfig.Hosts,
			"topic", kafkaConfig.Topic,
			"timeout", timeout,
			"group_id", groupID,
			"start_offset", kafkaConfig.StartOffset,
			"required_acks", kafkaConfig.RequiredAcks,
		)
	}

	return nil
}

func (k *kafkaService) KafkaClose() {
	if k.ChatWriter != nil {
		if err := k.ChatWriter.Close(); err != nil {
			if k.logger != nil {
				k.logger.Log(log.LevelError, "msg", "failed to close kafka writer", "err", err)
			}
		}
	}
	if k.ChatReader != nil {
		if err := k.ChatReader.Close(); err != nil {
			if k.logger != nil {
				k.logger.Log(log.LevelError, "msg", "failed to close kafka reader", "err", err)
			}
		}
	}
}

// CreateTopic 创建topic
func (k *kafkaService) CreateTopic(kafkaConfig *conf.Data_Kafka) error {
	if kafkaConfig == nil {
		return fmt.Errorf("kafka config is nil")
	}

	// 从配置读取分区数，默认 1
	partition := int(kafkaConfig.Partition)
	if partition <= 0 {
		partition = 1
	}

	// 从配置读取副本因子，默认 1
	replicationFactor := int(kafkaConfig.ReplicationFactor)
	if replicationFactor <= 0 {
		replicationFactor = 1
	}

	// 解析 hosts，取第一个地址用于连接
	hosts := strings.Split(kafkaConfig.Hosts, ",")
	if len(hosts) == 0 {
		return fmt.Errorf("kafka hosts is empty")
	}

	// 连接至任意kafka节点
	var err error
	// 建立与broker的tcp连接
	k.KafkaConn, err = kafka.Dial("tcp", hosts[0])
	if err != nil {
		if k.logger != nil {
			k.logger.Log(log.LevelError, "msg", "failed to dial kafka", "err", err)
		}
		return err
	}
	defer k.KafkaConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             kafkaConfig.Topic,
			NumPartitions:     partition,         // 分区数
			ReplicationFactor: replicationFactor, // 副本因子
		},
	}

	// 创建topic
	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		if k.logger != nil {
			k.logger.Log(log.LevelError, "msg", "failed to create kafka topic", "err", err)
		}
		return err
	}

	if k.logger != nil {
		k.logger.Log(log.LevelInfo,
			"msg", "kafka topic created",
			"topic", kafkaConfig.Topic,
			"partitions", partition,
			"replication_factor", replicationFactor,
		)
	}

	return nil
}
