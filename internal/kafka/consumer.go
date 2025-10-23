package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-kratos/kratos/v2/log"
	"strings"
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
		log.Debug("Kafka consumer not initialized, skipping consumer")
		return
	}

	partitions, err := consumer.Partitions(topic)
	log.Debug("Get all kafka partitions:", partitions)
	if err != nil {
		log.Debug("Failed to get partitions:", err)
		return
	}

	// 实时消费指定 Topic 中第 0 号分区的消息，并通过回调函数处理每一条消息。
	// sarama.OffsetNewest从最新消息开始消费
	//partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	//if nil != err {
	//	log.Debug("iConsumePartition error:", err)
	//	return
	//}
	//
	//defer partitionConsumer.Close()
	//for {
	//	msg := <-partitionConsumer.Messages()
	//	if nil != callBack {
	//		callBack(msg.Value)
	//	}
	//}

	for _, p := range partitions {
		pc, err := consumer.ConsumePartition(topic, p, sarama.OffsetNewest)
		if err != nil {
			log.Debug("ConsumePartition error:", err)
			continue
		}

		// 启动独立 goroutine 处理每个分区
		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for msg := range pc.Messages() {
				log.Debugf("[Partition %d] offset=%d", msg.Partition, msg.Offset)
				if callBack != nil {
					callBack(msg.Value)
				}
			}
		}(pc)
	}
}

func CloseConsumer() {
	if consumer != nil {
		consumer.Close()
	}
}
