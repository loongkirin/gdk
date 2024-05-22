package rocketmq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type rocketmqProducer struct {
	config   Config
	producer rocketmq.Producer
}

// NewProducer create new kafka producer
func NewProducer(cfg Config) *rocketmqProducer {
	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver(cfg.Brokers)),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil
	}
	err = p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		return nil
	}

	return &rocketmqProducer{config: cfg, producer: p}
}

func (p *rocketmqProducer) PublishMessage(ctx context.Context, msg string) error {
	rocketmqMsg := &primitive.Message{
		Topic: p.config.TopicName,
		Body:  []byte(msg),
	}
	_, err := p.producer.SendSync(ctx, rocketmqMsg)

	if err != nil {
		fmt.Printf("send message error: %s\n", err)
	}
	return err
}

func (p *rocketmqProducer) PublishMessageAsync(ctx context.Context, msg string) error {
	err := p.producer.SendAsync(ctx,
		func(ctx context.Context, result *primitive.SendResult, e error) {
			if e != nil {
				fmt.Printf("receive message error: %s\n", e)
			}
		}, primitive.NewMessage(p.config.TopicName, []byte(msg)))

	if err != nil {
		fmt.Printf("send message error: %s\n", err)
	}
	return err
}

func (p *rocketmqProducer) Close() {
	p.producer.Shutdown()
}
