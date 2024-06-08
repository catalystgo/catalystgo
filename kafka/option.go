package kafka

import (
	"time"

	"github.com/IBM/sarama"
)

// Sarama Options

type SaramaOption func(*sarama.Config)

// Consumer Group Options

type ConsumerGroupOption func(*sarama.Config)

func WithRebalanceStrategy(strategy sarama.BalanceStrategy) ConsumerGroupOption {
	return func(c *sarama.Config) {
		c.Consumer.Group.Rebalance.Strategy = strategy
	}
}

func WithNewestOffset() ConsumerGroupOption {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.Initial = sarama.OffsetNewest
	}
}

func WithOldestOffset() ConsumerGroupOption {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
}

// Producer Options

// Message Options

type MessageOption func(*sarama.ProducerMessage)

func WithKey(key string) MessageOption {
	return func(p *sarama.ProducerMessage) {
		p.Key = sarama.StringEncoder(key)
	}
}

func WithPartition(partition int32) MessageOption {
	return func(p *sarama.ProducerMessage) {
		p.Partition = partition
	}
}

func WithHeaders(headers []sarama.RecordHeader) MessageOption {
	return func(p *sarama.ProducerMessage) {
		p.Headers = headers
	}
}

func WithTimestamp(timestamp int64) MessageOption {
	return func(p *sarama.ProducerMessage) {
		p.Timestamp = time.Unix(timestamp, 0)
	}
}
