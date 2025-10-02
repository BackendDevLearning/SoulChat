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

func TestKafka(t *testing.T) {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	cf := sarama.NewConfig()
	client, _ := sarama.NewClient(strings.Split(bc.Data.Kafka.Hosts, ","), cf)

	producer, _ := sarama.NewSyncProducerFromClient(client)

	defer func() {
		if producer != nil {
			producer.Close()
		}
		if client != nil {
			client.Close()
		}
	}()

	var str string = "test"
	var data []byte = []byte(str)
	be := sarama.ByteEncoder(data)
	fmt.Println(be)
	producer.Input() <- &sarama.ProducerMessage{Topic: "test", Key: nil, Value: be}
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
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	cf := sarama.NewConfig()
	client, _ := sarama.NewClient(strings.Split(bc.Data.Kafka.Hosts, ","), cf)

	consumer, _ = sarama.NewConsumerFromClient(client)

	partitionConsumer, _ := consumer.ConsumePartition("test", 0, sarama.OffsetNewest)

	defer partitionConsumer.Close()
	for {
		msg := <-partitionConsumer.Messages()
		fmt.Println(msg.Value)
		fmt.Println(string(msg.Value))
	}
}
