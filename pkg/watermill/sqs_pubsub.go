package watermill

import (
	"github.com/ThreeDotsLabs/watermill-amazonsqs/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
	"go.uber.org/zap"
)

func getAWSConfig(sqsConfig *config.SQSConfig) aws.Config {
	return aws.Config{
		Region: &sqsConfig.AWSRegion,
	}
}

// NewSQSPublisher creates a Publisher.
func NewSQSPublisher(sqsConfig *config.SQSConfig) (message.Publisher, error) {
	cfg := getAWSConfig(sqsConfig)

	publisherCfg := sqs.PublisherConfig{
		AWSConfig: cfg,
		Marshaler: sqs.DefaultMarshalerUnmarshaler{},
	}

	publisher, err := sqs.NewPublisher(
		publisherCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	if err == nil {
		logger.Info("SQS publisher connected successfully",
			zap.Any("region", sqsConfig.AWSRegion),
		)
	}

	return publisher, err
}

// NewSQSSubscriber creates a Subscriber.
func NewSQSSubscriber(sqsConfig *config.SQSConfig) (message.Subscriber, error) {
	cfg := getAWSConfig(sqsConfig)

	subscriberCfg := sqs.SubscriberConfig{
		AWSConfig:   cfg,
		Unmarshaler: sqs.DefaultMarshalerUnmarshaler{},
	}

	subscriber, err := sqs.NewSubsciber(
		subscriberCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	if err == nil {
		logger.Info("SQS subscriber connected successfully",
			zap.Any("region", sqsConfig.AWSRegion),
		)
	}

	return subscriber, err
}
