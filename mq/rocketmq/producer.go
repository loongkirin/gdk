package rocketmq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

type rocketmqProducer struct {
	config   Config
	producer golang.Producer
}

// NewProducer create new kafka producer
func NewProducer(cfg Config) *rocketmqProducer {
	p, err := golang.NewProducer(&golang.Config{
		Endpoint:      cfg.Endpoint,
		ConsumerGroup: cfg.GroupName,
		NameSpace:     cfg.NameSpace,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    cfg.AccessKey,
			AccessSecret: cfg.SecretKey,
		},
	},
		golang.WithTopics(cfg.TopicName),
	)
	if err != nil {
		fmt.Printf("create producer error: %s", err.Error())
		return nil
	}
	err = p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		return nil
	}
	return &rocketmqProducer{
		config:   cfg,
		producer: p,
	}
}

func (p *rocketmqProducer) PublishMessage(ctx context.Context, msg string) error {
	rocketmqMsg := &golang.Message{
		Topic: p.config.TopicName,
		Body:  []byte(msg),
	}
	_, err := p.producer.Send(ctx, rocketmqMsg)

	if err != nil {
		fmt.Printf("send message error: %s\n", err)
	}
	return err
}

func (p *rocketmqProducer) PublishMessageAsync(ctx context.Context, msg string) error {
	rocketmqMsg := &golang.Message{
		Topic: p.config.TopicName,
		Body:  []byte(msg),
	}
	// set keys and tag
	// rocketmqMsg.SetKeys("a", "b")
	// rocketmqMsg.SetTag("ab")
	// send message in async
	p.producer.SendAsync(context.TODO(), rocketmqMsg, func(ctx context.Context, resp []*golang.SendReceipt, err error) {
		if err != nil {
			fmt.Printf("send message error: %s\n", err)
		}
	})
	return nil
}

func (p *rocketmqProducer) Close() {
	p.producer.GracefulStop()
}
