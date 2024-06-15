package realtime

import (
	"context"
)

type Client interface {
	GetValue(ctx context.Context, key string) (Value, error)
	SetValue(ctx context.Context, key string, value string) error
	Watch(ctx context.Context, key string, cb func(old Value, new Value))
	Close() error
}
