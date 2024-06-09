package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, address string) (*redis.Client, error) {
	opt, err := redis.ParseURL(address)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return client, nil
}
