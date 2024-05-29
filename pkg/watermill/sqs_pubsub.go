package watermill

import (
	"context"

	"github.com/ThreeDotsLabs/watermill-amazonsqs/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	cfg "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/logger"
	"go.uber.org/zap"
)

// NewSQSPublisher creates a Publisher.
func NewSQSPublisher(awsConfig *cfg.AWSConfig) (message.Publisher, error) {
	sdkConfig, err := aws_repository.GetAWSConfig(context.Background(), awsConfig)
	if err != nil {
		return nil, err
	}

	publisherCfg := sqs.PublisherConfig{
		AWSConfig: sdkConfig,
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
func NewSQSSubscriber(awsConfig *cfg.AWSConfig) (message.Subscriber, error) {
	sdkConfig, err := aws_repository.GetAWSConfig(context.Background(), awsConfig)
	if err != nil {
		return nil, err
	}

	subscriberCfg := sqs.SubscriberConfig{
		AWSConfig:   sdkConfig,
		Unmarshaler: sqs.DefaultMarshalerUnmarshaler{},
	}

	subscriber, err := sqs.NewSubscriber(
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
