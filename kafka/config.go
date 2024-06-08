package kafka

import "github.com/IBM/sarama"

type ConfigOption func(*Config)

func WithClientID(clientID string) ConfigOption {
	return func(c *Config) {
		c.ClientID = clientID
	}
}

func WithVersion(version sarama.KafkaVersion) ConfigOption {
	return func(c *Config) {
		c.Version = version
	}
}

type Config struct {
	ClientID string
	Brokers  []string
	Version  sarama.KafkaVersion

	Opts []func(*sarama.Config)
}

func NewConfig(brokers []string, opts ...func(*Config)) *Config {
	c := &Config{
		ClientID: "CatalystGo", // default client ID
		Brokers:  brokers,
		Version:  sarama.V2_6_0_0, // default version
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func NewSyncProducerConfig(brokers []string, opts ...func(*Config)) *Config {
	syncProducerConfig := []func(*sarama.Config){
		func(c *sarama.Config) {
			c.Producer.Idempotent = true
			c.Producer.Return.Errors = true
			c.Producer.RequiredAcks = sarama.WaitForAll
		},
	}

	cfg := NewConfig(brokers, opts...)
	cfg.Opts = append(cfg.Opts, syncProducerConfig...)

	return cfg
}

func NewAsyncProducerConfig(brokers []string, opts ...func(*Config)) *Config {
	asyncProducerConfig := []func(*sarama.Config){
		func(c *sarama.Config) {
			c.Producer.Return.Errors = true
		},
	}

	cfg := NewConfig(brokers, opts...)
	cfg.Opts = append(cfg.Opts, asyncProducerConfig...)

	return cfg
}
