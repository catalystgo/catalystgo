package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/catalystgo/tracerok/logger"
	"go.uber.org/zap"
)

type SyncProducer interface {
	Produce(topic string, message []byte, opts ...MessageOption) error
	ProduceWithContext(ctx context.Context, topic string, message []byte, opts ...MessageOption) error
	Close() error
}

type syncProducer struct {
	sarama.SyncProducer
}

func NewSyncProducer(cfg *Config) (SyncProducer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = cfg.Version
	saramaConfig.ClientID = cfg.ClientID
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll

	for _, opt := range cfg.Opts {
		opt(saramaConfig)
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &syncProducer{SyncProducer: producer}, nil
}

func (p *syncProducer) Produce(topic string, message []byte, opts ...MessageOption) error {
	return p.ProduceWithContext(context.Background(), topic, message, opts...)
}

func (p *syncProducer) ProduceWithContext(ctx context.Context, topic string, message []byte, opts ...MessageOption) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	for _, opt := range opts {
		opt(msg)
	}

	return p.producer(ctx, topic, msg)
}

func (p *syncProducer) producer(ctx context.Context, topic string, msg *sarama.ProducerMessage) error {
	partition, offset, err := p.SyncProducer.SendMessage(msg)
	if err != nil {
		logger.ErrorKV(ctx, "syncProducer send message error",
			zap.Error(err),
			zap.String("topic", topic),
			zap.Int64("offset", offset),
			zap.Int32("partition", partition),
		)

		return err
	}

	return nil
}

func (p *syncProducer) Close() error {
	return p.SyncProducer.Close()
}
