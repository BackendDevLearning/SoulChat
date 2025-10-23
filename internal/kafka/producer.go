package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"strings"
)

// Sarama：Go 语言官方常用的 Kafka 客户端库。
var producer sarama.AsyncProducer

// 消息订阅和分发的时候必须指定 topic
var topic string = "default_message"

func InitProducer(topicInput, hosts string) error {
	topic = topicInput // 当前服务kafka 的 topic
	//hosts 当前集群的地址列表，多个用逗号隔开
	config := sarama.NewConfig()
	// 压缩方式	含义
	//sarama.CompressionNone	不压缩（默认）。
	//sarama.CompressionGZIP	GZIP 压缩，压缩率高，适合大消息。
	//sarama.CompressionSnappy	Snappy 压缩，速度快，压缩率一般。
	//sarama.CompressionLZ4	LZ4 压缩，速度和压缩率平衡。
	config.Producer.Compression = sarama.CompressionGZIP
	//连接到 Kafka 集群。
	//发现所有 Broker 节点。
	//建立必要的元数据缓存
	client, err := sarama.NewClient(strings.Split(hosts, ","), config)
	if nil != err {
		fmt.Println("init kafka client error", err.Error())
		return err
	}

	// 异步生产者，消息发送是非阻塞的
	producer, err = sarama.NewAsyncProducerFromClient(client)
	if nil != err {
		fmt.Println("init kafka async client error", err.Error())
		return err
	}
	return nil
}

func Send(data []byte) {
	if producer == nil {
		fmt.Println("Kafka producer not initialized, skipping message")
		return
	}
	be := sarama.ByteEncoder(data)
	// 把二进制消息发到 kafka
	producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   nil, // 默认采用轮询
		Value: be,
	}
}

func Close() {
	if producer != nil {
		producer.Close()
	}
}
