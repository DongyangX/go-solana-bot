package utils

import (
	"context"
	"log"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type MqUtil struct {
	producer rocketmq.Producer
}

func NewMqUtil(url string) *MqUtil {
	// 创建 Producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{url}), // 和 consumer 一致
		producer.WithRetry(2),
	)
	if err != nil {
		panic(err)
	}

	// 启动 Producer
	err = p.Start()
	if err != nil {
		panic(err)
	}
	return &MqUtil{
		producer: p,
	}
}

func (mq *MqUtil) Stop() {
	_ = mq.producer.Shutdown()
}

func (mq *MqUtil) Send(topic string, message []byte) (*primitive.SendResult, error) {
	// 构造消息
	msg := &primitive.Message{
		Topic: topic,
		Body:  message,
	}

	// 同步发送消息
	res, err := mq.producer.SendSync(context.Background(), msg)
	if err != nil {
		log.Println("Send failed:", err)
	} else {
		log.Println("Send success:", res.String())
	}
	return res, err
}
