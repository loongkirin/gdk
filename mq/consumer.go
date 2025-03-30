package mq

import (
	"context"
)

type Message any

type ConsumeMessage func(ctx context.Context, msg Message) error

type Consumer interface {
	Start(ctx context.Context, fn ConsumeMessage) error
	Close()
}
