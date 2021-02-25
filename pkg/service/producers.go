package service

import (
	"github.com/moov-io/base/log"
	"github.com/moovfinancial/events/go/producer"
)

type ProducerConfig struct {
	Brokers []string
	TLS     bool

	Key    string
	Secret string
}

func NewProducer(logger log.Logger, config ProducerConfig, topic string) (producer.Producer, error) {

	configurable := []producer.ConfigModifier{
		producer.LogErrors(logger),
	}

	if config.Key != "" {
		configurable = append(configurable, producer.SASL(config.Key, config.Secret))
	}

	if config.TLS {
		configurable = append(configurable, producer.DefaultTLS())
	}

	p, err := producer.New(
		producer.ServiceName("ach-test-harness"),
		producer.Brokers(config.Brokers),
		producer.Topic(topic),
		producer.Consistent(),
		producer.LowLatency(),
		producer.All(),
		configurable...,
	)
	if err != nil {
		return nil, logger.Error().LogErrorf("unable to create producer - topic: %s - %w", topic, err).Err()
	}

	return p, nil
}
