package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/loongkirin/gdk/mq"
)

type consumer struct {
	config   Config
	consumer *kafka.Consumer
	sigint   chan bool
}

func NewConsumer(cfg Config) *consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		"broker.address.family": "v4",
		"group.id":              cfg.GroupID,
		"session.timeout.ms":    6000,
		// Start reading from the first message of each assigned
		// partition if there are no previously committed offsets
		// for this group.
		"auto.offset.reset": "earliest",
		// Whether or not we store offsets automatically.
		"enable.auto.offset.store": false,
	})

	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		return nil
	}

	return &consumer{
		config:   cfg,
		consumer: c,
	}
}

func (c *consumer) Start(ctx context.Context, fn mq.ConsumeMessage) error {
	c.sigint = make(chan bool, 1)
	err := c.consumer.Subscribe(c.config.TopicName, nil)
	if err != nil {
		return err
	}

	running := true
	for running {
		select {
		case sig := <-c.sigint:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			running = false
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				// Process the message received.
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}

				err = fn(ctx, e)
				if err != nil {
					continue
				}

				// We can store the offsets of the messages manually or let
				// the library do it automatically based on the setting
				// enable.auto.offset.store. Once an offset is stored, the
				// library takes care of periodically committing it to the broker
				// if enable.auto.commit isn't set to false (the default is true).
				// By storing the offsets manually after completely processing
				// each message, we can ensure atleast once processing.
				_, err := c.consumer.StoreMessage(e)
				if err != nil {
					fmt.Printf("Error storing offset after message %s:\n",
						e.TopicPartition)
				}
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				fmt.Printf("Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					running = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
	return nil
}

func (c *consumer) Close() {
	if c.consumer.IsClosed() {
		return
	}
	c.sigint <- false
	close(c.sigint)
	c.consumer.Unsubscribe()
	c.consumer.Close()
}
