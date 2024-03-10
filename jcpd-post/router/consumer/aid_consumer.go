package consumer

import (
	"github.com/IBM/sarama"
)

type Consumer struct{}

// Setup 在消费者开始消费消息前调用的方法
func (consumer *Consumer) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup 在消费者停止消费消息后调用的方法
func (consumer *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim 消费者实际消费消息的方法
func (consumer *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	flag := 0
	for msg := range claim.Messages() {
		KafkaListener.work()
		//	过滤掉每次启动的第一个, 因为会重复 ...
		if flag == 0 {
			flag++
			continue
		}
		// 进行消息的处理

		// 手动标记偏移量并提交, 防止消息丢失和重复消费
		sess.MarkOffset(msg.Topic, msg.Partition, msg.Offset, "")
		sess.Commit()
		KafkaListener.rest()
	}
	return nil
}
