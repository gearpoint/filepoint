package watermill

import (
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
)

const (
	// defines the Key used to define the partition.
	KafkaKey = "partition"
)

// NewKafkaPublisher creates a Publisher.
func NewKafkaPublisher(kafkaConfig *config.KafkaConfig) (message.Publisher, error) {
	saramaPublisherCfg := kafka.DefaultSaramaSyncPublisherConfig()
	saramaPublisherCfg.Producer.MaxMessageBytes = kafkaConfig.MaxMessageBytes
	saramaPublisherCfg.Producer.Retry.Max = kafkaConfig.MaxRetries

	publisherCfg := kafka.PublisherConfig{
		Brokers:               kafkaConfig.Brokers,
		Marshaler:             getKafkaMarshaler(),
		OverwriteSaramaConfig: saramaPublisherCfg,
		Tracer:                nil, // verify otel
	}

	publisher, err := kafka.NewPublisher(
		publisherCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	return publisher, err
}

// NewKafkaSubscriber creates a Subscriber.
func NewKafkaSubscriber(kafkaConfig *config.KafkaConfig) (message.Subscriber, error) {
	saramaPublisherCfg := kafka.DefaultSaramaSyncPublisherConfig()
	saramaPublisherCfg.Consumer.Offsets.Retry.Max = kafkaConfig.MaxRetries

	subscriberCfg := kafka.SubscriberConfig{
		Brokers:               kafkaConfig.Brokers,
		Unmarshaler:           getKafkaMarshaler(),
		OverwriteSaramaConfig: saramaPublisherCfg,
		Tracer:                nil,
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
			return msg.Metadata.Get(KafkaKey), nil
		},
	)
}
