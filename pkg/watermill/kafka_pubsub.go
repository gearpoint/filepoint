package watermill

import (
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
)

const (
	// defines the Key used to define the partition.
	KafkaPartitionKey = "partition"
)

// NewKafkaPublisher creates a Publisher.
func NewKafkaPublisher(kafkaConfig *config.KafkaConfig) (message.Publisher, error) {
	saramaCfg := kafka.DefaultSaramaSyncPublisherConfig()
	saramaCfg.Producer.MaxMessageBytes = kafkaConfig.MaxMessageBytes
	saramaCfg.Producer.Retry.Max = kafkaConfig.MaxRetries

	publisherCfg := kafka.PublisherConfig{
		Brokers:               kafkaConfig.Brokers,
		Marshaler:             getKafkaMarshaler(),
		OverwriteSaramaConfig: saramaCfg,
		Tracer:                nil, // todo: verify otel
	}

	publisher, err := kafka.NewPublisher(
		publisherCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	return publisher, err
}

// NewKafkaSubscriber creates a Subscriber.
func NewKafkaSubscriber(kafkaConfig *config.KafkaConfig) (message.Subscriber, error) {
	saramaCfg := kafka.DefaultSaramaSyncPublisherConfig()
	saramaCfg.Consumer.Offsets.Retry.Max = kafkaConfig.MaxRetries

	subscriberCfg := kafka.SubscriberConfig{
		Brokers:               kafkaConfig.Brokers,
		Unmarshaler:           getKafkaMarshaler(),
		OverwriteSaramaConfig: saramaCfg,
		Tracer:                nil, // todo: verify otel
	}

	subscriber, err := kafka.NewSubscriber(
		subscriberCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	return subscriber, err
}

// getKafkaMarshaler returns the configured marshaler.
func getKafkaMarshaler() kafka.MarshalerUnmarshaler {
	return kafka.NewWithPartitioningMarshaler(
		func(topic string, msg *message.Message) (string, error) {
			return msg.Metadata.Get(KafkaPartitionKey), nil
		},
	)
}
