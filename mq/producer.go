package mq

import (
	"context"
)

type Producer interface {
	PublishMessage(ctx context.Context, msg string) error
	PublishMessageAsync(ctx context.Context, msg string) error
	Close()
}
