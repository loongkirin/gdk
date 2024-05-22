package rocketmq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/loongkirin/gdk/mq"
)

type rocketmqConsumer struct {
	config   Config
	consumer rocketmq.PushConsumer
	sigint   chan bool
}

func NewConsumer(cfg Config) *rocketmqConsumer {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(cfg.GroupName),
		consumer.WithNsResolver(primitive.NewPassthroughResolver(cfg.Brokers)),
	)

	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		return nil
	}

	return &rocketmqConsumer{
		config:   cfg,
		consumer: c,
	}
}

func (c *rocketmqConsumer) Start(ctx context.Context, fn mq.ConsumeMessage) error {
	c.sigint = make(chan bool, 1)
	err := c.consumer.Subscribe(c.config.TopicName, consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			err := fn(ctx, msgs[i])
			if err != nil {
				fmt.Printf("Failed to consumer message: %s\n", err)
				return consumer.ConsumeRetryLater, err
			}
		}

		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = c.consumer.Start()
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func (c *rocketmqConsumer) Close() {
	c.sigint <- false
	close(c.sigint)
	c.consumer.Unsubscribe(c.config.TopicName)
	c.consumer.Shutdown()
}
