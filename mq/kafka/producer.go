package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/loongkirin/gdk/util"
)

type producer struct {
	config   Config
	producer *kafka.Producer
}

// NewProducer create new kafka producer
func NewProducer(cfg Config) *producer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": cfg.Brokers[0]})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		return nil
	}
	return &producer{config: cfg, producer: p}
}

func (p *producer) PublishMessage(ctx context.Context, msg string) error {
	for {
		kafkaMsg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &p.config.TopicName, Partition: int32(p.config.Partition)},
			Value:          []byte(msg),
			Key:            []byte(util.GenerateId()),
			Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
			Timestamp:      time.Now().UTC(),
			TimestampType:  kafka.TimestampCreateTime,
		}
		err := p.producer.Produce(kafkaMsg, nil)

		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrQueueFull {
				// Producer queue is full, wait 1s for messages
				// to be delivered then try again.
				time.Sleep(time.Second * 3)
				continue
			}
			fmt.Printf("Failed to produce message: %v\n", err)
			return err
		}
		return nil
	}
}

func (p *producer) PublishMessageAsync(ctx context.Context, msg string) error {
	c := make(chan error)
	go func(msg string, c chan<- error) {
		err := p.PublishMessage(ctx, msg)
		c <- err
	}(msg, c)

	err := <-c
	close(c)
	return err
}

func (p *producer) Close() {
	if p.producer.IsClosed() {
		return
	}

	for p.producer.Flush(10000) > 0 {
		fmt.Print("Still waiting to flush outstanding messages\n")
	}
	p.producer.Close()
}
