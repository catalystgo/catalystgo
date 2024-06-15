package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/catalystgo/logger/logger"
	"go.uber.org/zap"
)

type Handler func(message *sarama.ConsumerMessage) error

type ConsumerGroup interface {
	Consume(ctx context.Context, topic []string, handler Handler) error
	Close() error
}

type consumerGroup struct {
	cfg     *Config
	groupID string

	sarama.ConsumerGroup
}

func NewConsumerGroup(cfg *Config, groupID string) (ConsumerGroup, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = cfg.Version
	saramaConfig.ClientID = cfg.ClientID
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategySticky()

	for _, opt := range cfg.Opts {
		opt(saramaConfig)
	}

	conumer, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	cg := &consumerGroup{
		cfg:           cfg,
		groupID:       groupID,
		ConsumerGroup: conumer,
	}

	return cg, nil
}

func (cg *consumerGroup) Consume(ctx context.Context, topic []string, handler Handler) error {
	h := &consumerGroupHandler{handler: handler}
	go func() {
		for {
			if err := cg.ConsumerGroup.Consume(ctx, topic, h); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					logger.ErrorKV(ctx, fmt.Sprintf("consumer group %s closed", cg.groupID),
						zap.Error(err),
						zap.Strings("topics", topic),
					)

					return
				}

				logger.Errorf(context.Background(), "error consuming: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	go cg.handleErrors()

	return nil
}

func (cg *consumerGroup) Close() error {
	return cg.ConsumerGroup.Close()
}

func (cg *consumerGroup) handleErrors() {
	for err := range cg.ConsumerGroup.Errors() {
		logger.Errorf(context.Background(), "consumer group error: %v", err)
	}
}

type consumerGroupHandler struct {
	handler Handler
}

func (h *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	logger.WarnKV(context.Background(), "consumer group setup initialized",
		zap.String("member_id", session.MemberID()),
		zap.Strings("topics", getSessionTopics(session)),
	)
	return nil
}

func (h *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	logger.WarnKV(context.Background(), "consumer group cleanup initialized",
		zap.String("member_id", session.MemberID()),
		zap.Strings("topics", getSessionTopics(session)),
	)
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for message := range claim.Messages() {
		if err := h.handler(message); err != nil {
			logger.ErrorKV(ctx, fmt.Sprintf("consumer group handle message error: %v", err),
				zap.String("topic", message.Topic),
				zap.Int64("offset", message.Offset),
				zap.Int32("partition", message.Partition),
			)
		}

		session.MarkMessage(message, "")
	}
	return nil
}

// getSessionTopics returns the topics that the session is consuming
func getSessionTopics(session sarama.ConsumerGroupSession) []string {
	topics := make([]string, 0, len(session.Claims()))
	for topic := range session.Claims() {
		topics = append(topics, topic)
	}
	return topics
}
