package kafka

import (
	"context"
	"fmt"
	"kratos-realworld/internal/conf"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var ctx = context.Background()

type kafkaService struct {
	ChatWriter *kafka.Writer
	ChatReader *kafka.Reader
	KafkaConn  *kafka.Conn
	logger     *zap.Logger
}

var KafkaService = new(kafkaService)

// KafkaInit 初始化kafka
func (k *kafkaService) KafkaInit(kafkaConfig *conf.Data_Kafka, logger *zap.Logger, timeout int, partition int) error {
	k.logger = logger
	if kafkaConfig == nil {
		return fmt.Errorf("kafka config is nil")
	}

	// 解析 hosts，支持逗号分隔的多个地址
	hosts := strings.Split(kafkaConfig.Hosts, ",")
	if len(hosts) == 0 {
		return fmt.Errorf("kafka hosts is empty")
	}

	// 默认超时时间为 10 秒
	if timeout <= 0 {
		timeout = 10
	}

	// 默认分区数为 1
	if partition <= 0 {
		partition = 1
	}

	// 创建 Writer
	k.ChatWriter = &kafka.Writer{
		Addr:                   kafka.TCP(hosts...),
		Topic:                  kafkaConfig.Topic,
		// 分区策略，负载均衡
		Balancer:               &kafka.Hash{},
		// 写入超时时间
		WriteTimeout:           time.Duration(timeout) * time.Second,
		// 消息确认机制，RequireNone表示不需要确认
		RequiredAcks:           kafka.RequireNone,
		// 是否允许自动创建主题
		AllowAutoTopicCreation: false,
	}

	// 创建 Reader
	k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        hosts,
		Topic:          kafkaConfig.Topic,
		// 提交偏移量（Offset）的时间间隔，相当于提交消费的进度
		CommitInterval: time.Duration(timeout) * time.Second,
		// 同一 GroupID 的消费者共享消费进度，相当于共享消费的进度
		// 用于负载均衡和故障恢复，当一个消费者宕机时，其他消费者可以继续消费
		GroupID:     "chat",
		// 首次启动时的起始消费位置，从最新消息开始消费（不消费历史消息）
		StartOffset: kafka.LastOffset,
	})

	if logger != nil {
		logger.Info("kafka service initialized",
			zap.String("hosts", kafkaConfig.Hosts),
			zap.String("topic", kafkaConfig.Topic),
		)
	}

	return nil
}

func (k *kafkaService) KafkaClose() {
	if k.ChatWriter != nil {
		if err := k.ChatWriter.Close(); err != nil {
			if k.logger != nil {
				k.logger.Error("failed to close kafka writer", zap.Error(err))
			}
		}
	}
	if k.ChatReader != nil {
		if err := k.ChatReader.Close(); err != nil {
			if k.logger != nil {
				k.logger.Error("failed to close kafka reader", zap.Error(err))
			}
		}
	}
}

// CreateTopic 创建topic
func (k *kafkaService) CreateTopic(kafkaConfig *conf.Data_Kafka, partition int) error {
	if kafkaConfig == nil {
		return fmt.Errorf("kafka config is nil")
	}

	// 默认分区数为 1
	if partition <= 0 {
		partition = 1
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
			k.logger.Error("failed to dial kafka", zap.Error(err))
		}
		return err
	}
	defer k.KafkaConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             kafkaConfig.Topic,
			NumPartitions:     partition, // 分区数
			ReplicationFactor: 1,
		},
	}

	// 创建topic
	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		if k.logger != nil {
			k.logger.Error("failed to create kafka topic", zap.Error(err))
		}
		return err
	}

	if k.logger != nil {
		k.logger.Info("kafka topic created",
			zap.String("topic", kafkaConfig.Topic),
			zap.Int("partitions", partition),
		)
	}

	return nil
}
