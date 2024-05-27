package watermill

import (
	"github.com/ThreeDotsLabs/watermill-amazonsqs/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/logger"
	"go.uber.org/zap"
)

func getAWSConfig(awsConfig *config.AWSConfig) aws.Config {
	return aws.Config{
		Endpoint: &awsConfig.Endpoint,
		Region:   &awsConfig.Region,
	}
}

// NewSQSPublisher creates a Publisher.
func NewSQSPublisher(awsConfig *config.AWSConfig) (message.Publisher, error) {
	cfg := getAWSConfig(awsConfig)

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
			zap.Any("region", awsConfig.Region),
		)
	}

	return publisher, err
}

// NewSQSSubscriber creates a Subscriber.
func NewSQSSubscriber(awsConfig *config.AWSConfig) (message.Subscriber, error) {
	cfg := getAWSConfig(awsConfig)

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
			zap.Any("region", awsConfig.Region),
		)
	}

	return subscriber, err
}
