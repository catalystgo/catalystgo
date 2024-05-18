package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/catalystgo/catalystgo/errors"
	"github.com/catalystgo/catalystgo/internal/realtime"
	"github.com/catalystgo/tracerok/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

var (
	errNoValue = errors.New("no value found").Code(errors.NotFound)
)

type Option func(*clientv3.Config)

func WithTLS(tlsConfig *tls.Config) Option {
	return func(c *clientv3.Config) {
		c.TLS = tlsConfig
	}
}

type Client struct {
	client *clientv3.Client

	ctx    context.Context
	cancel func()
}

func NewClient(ctx context.Context, endpoints []string, opts ...Option) (*Client, error) {
	ctx, cancel := context.WithCancel(ctx)

	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		Context:     ctx,
	}

	for _, o := range opts {
		o(&config)
	}

	client, err := clientv3.New(config)

	if err != nil {
		cancel()
		return nil, errors.Newf("create etcd client: %v", err)
	}

	return &Client{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (c *Client) GetValue(ctx context.Context, key string) (realtime.Value, error) {
	response, err := c.client.Get(ctx, key, clientv3.WithLimit(1))
	if err != nil {
		return nil, errors.Newf("get etcd value by key: %s => %v", key, err)
	}

	if len(response.Kvs) == 0 {
		logger.Warnf(ctx, "no value found for key \"%s\"", key)
		return realtime.NewNilValue(errNoValue), nil
	}

	// Return the first element in the array of keys
	value := realtime.NewValue(string(response.Kvs[0].Value))

	return value, nil
}

// SetValue sets a value by key in the etcd storage
func (c *Client) SetValue(ctx context.Context, key string, value string) error {
	response, err := c.client.Put(ctx, key, value)
	if err != nil {
		return errors.Newf("put value for etcd key: %s => %v", key, err)
	}

	logger.WarnKV(ctx, fmt.Sprintf("set value for key \"%s\" ", key),
		zap.String("key", key),
		zap.String("value", value),

		zap.Int64("etcd_revision", response.Header.GetRevision()),
		zap.Uint64("etcd_raft_term", response.Header.GetRaftTerm()),
		zap.Uint64("etcd_cluster_id", response.Header.GetClusterId()),
	)

	return nil
}

// Watch watches a key and calls `cb` upon change events
func (c *Client) Watch(ctx context.Context, key string, cb func(old realtime.Value, new realtime.Value)) {
	response := c.client.Watcher.Watch(ctx, key, clientv3.WithPrevKV())
	go func(ch clientv3.WatchChan) {
		for {
			select {
			case event := <-ch:
				for _, e := range event.Events {
					var oldValue, newValue string
					if e.PrevKv != nil {
						oldValue = string(e.PrevKv.Value)
					}
					if e.Kv != nil {
						newValue = string(e.Kv.Value)
					}

					logger.WarnKV(ctx, fmt.Sprintf("value of key \"%s\" has changed", key),
						zap.String("event", e.Type.String()),
						zap.String("old", oldValue),
						zap.String("new", newValue),
					)

					cb(realtime.NewValue(oldValue), realtime.NewValue(newValue))
				}
			case err := <-c.ctx.Done():
				if errors.Is(c.ctx.Err(), context.Canceled) {
					logger.Warnf(ctx, "stopped watching key: %s => %v", key, err)
				}
				return
			}
		}
	}(response)
}

// Close closes etcd client connection
func (c *Client) Close() error {
	c.cancel()
	return c.client.Close()
}
