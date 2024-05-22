package rocketmq

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/loongkirin/gdk/mq"
)

var (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive messages in a loop
)

type rocketmqConsumer struct {
	config   Config
	consumer golang.SimpleConsumer
}

func NewConsumer(cfg Config) *rocketmqConsumer {
	c, err := golang.NewSimpleConsumer(&golang.Config{
		Endpoint:      cfg.Endpoint,
		ConsumerGroup: cfg.GroupName,
		NameSpace:     cfg.NameSpace,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    cfg.AccessKey,
			AccessSecret: cfg.SecretKey,
		},
	},
		golang.WithAwaitDuration(awaitDuration),
		golang.WithSubscriptionExpressions(map[string]*golang.FilterExpression{
			cfg.TopicName: golang.SUB_ALL,
		}),
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
	err := c.consumer.Start()
	if err != nil {
		fmt.Printf("Failed to start consumer: %s\n", err)
		return err
	}

	go func() {
		for {
			fmt.Println("start recevie message")
			mvs, err := c.consumer.Receive(ctx, maxMessageNum, invisibleDuration)
			if err != nil {
				fmt.Println(err)
			}
			// ack message
			for _, mv := range mvs {
				err = fn(ctx, mv)
				if err != nil {
					c.consumer.Ack(ctx, mv)
				}
			}
			time.Sleep(time.Second * 3)
		}
	}()
	return nil
}

func (c *rocketmqConsumer) Close() {
	c.consumer.Unsubscribe(c.config.TopicName)
	c.consumer.GracefulStop()
}
