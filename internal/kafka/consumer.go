package kafka

import (
	"github.com/IBM/sarama"
	"strings"
	"fmt"
)

var consumer sarama.Consumer

type ConsumerCallback func(data []byte)

// 初始化消费者
func InitConsumer(hosts string) error {
	config := sarama.NewConfig()
	client, err := sarama.NewClient(strings.Split(hosts, ","), config)
	if nil != err {
		fmt.Println("init kafka consumer client error", err.Error())
		return err
	}

	consumer, err = sarama.NewConsumerFromClient(client)
	if nil != err {
		fmt.Println("init kafka consumer error", err.Error())
		return err
	}
	return nil
}

// 消费消息，通过回调函数进行
func ConsumerMsg(callBack ConsumerCallback) {
	if consumer == nil {
		fmt.Println("Kafka consumer not initialized, skipping consumer")
		return
	}
	// 实时消费指定 Topic 中第 0 号分区的消息，并通过回调函数处理每一条消息。
	// sarama.OffsetNewest从最新消息开始消费
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if nil != err {
		fmt.Println("iConsumePartition error", err.Error())
		return
	}

	defer partitionConsumer.Close()
	for {
		msg := <-partitionConsumer.Messages()
		if nil != callBack {
			callBack(msg.Value)
		}
	}
}

func CloseConsumer() {
	if consumer != nil {
		consumer.Close()
	}
}
