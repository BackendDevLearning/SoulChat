package test

import (
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"kratos-realworld/internal/conf"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func TestKafkaProducer(t *testing.T) {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		t.Fatalf("Failed to scan config: %v", err)
	}

	if bc.Data.Kafka == nil {
		t.Skip("Kafka not configured")
	}

	cf := sarama.NewConfig()
	cf.Producer.Return.Successes = true
	cf.Producer.Timeout = 5 * time.Second

	fmt.Println("bc.Data.Kafka", bc.Data.Kafka)

	client, err := sarama.NewClient(strings.Split(bc.Data.Kafka.Hosts, ","), cf)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %v", err)
	}

	// 同步生产者，消息发送是阻塞的，会等待Kafka发送确认ack
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		t.Fatalf("Failed to create Kafka producer: %v", err)
	}

	defer func() {
		if producer != nil {
			producer.Close()
		}
		if client != nil {
			client.Close()
		}
	}()

	// 发送测试消息
	testMessage := "Hello Kafka from Go test!"
	message := &sarama.ProducerMessage{
		Topic: bc.Data.Kafka.Topic,
		Key:   sarama.StringEncoder("test-key"),
		Value: sarama.StringEncoder(testMessage),
	}

	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	fmt.Printf("Message sent successfully! Partition: %d, Offset: %d\n", partition, offset)
	t.Logf("Sent message: %s to topic: %s", testMessage, bc.Data.Kafka.Topic)
}

var consumer sarama.Consumer

type ConsumerCallBack func(data []byte)

func TestKafkaConsumer(t *testing.T) {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		t.Fatalf("Failed to scan config: %v", err)
	}

	if bc.Data.Kafka == nil {
		t.Skip("Kafka not configured")
	}

	cf := sarama.NewConfig()
	cf.Consumer.Return.Errors = true

	fmt.Println("bc.Data.Kafka", bc.Data.Kafka)

	client, err := sarama.NewClient(strings.Split(bc.Data.Kafka.Hosts, ","), cf)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %v", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		t.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	defer func() {
		if consumer != nil {
			consumer.Close()
		}
		if client != nil {
			client.Close()
		}
	}()

	// 消费消息（只消费一条用于测试）
	//partitionConsumer, err := consumer.ConsumePartition(bc.Data.Kafka.Topic, 0, sarama.OffsetNewest)
	// 消费上面生产者发送到Kafka里面的消息
	// sarama.OffsetOldest 从最早消息开始
	// sarama.OffsetNewest 从最新消息开始（之前的消息就不会被消费了），就是调用ConsumePartition时Partition里当前最大的offset + 1
	partitionConsumer, err := consumer.ConsumePartition(bc.Data.Kafka.Topic, 1, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("Failed to create partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	// 设置超时，避免无限等待
	timeout := time.After(10 * time.Second)

	select {
	case msg := <-partitionConsumer.Messages():
		fmt.Printf("Received message: %s\n", string(msg.Value))
		t.Logf("Received message from topic %s: %s", msg.Topic, string(msg.Value))
	case err := <-partitionConsumer.Errors():
		t.Fatalf("Consumer error: %v", err)
	case <-timeout:
		t.Log("No messages received within timeout period")
	}
}
