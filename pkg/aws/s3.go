package aws

import (
	"context"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gearpoint/filepoint/config"
)

func NewS3Client(awsCfg *config.S3) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)

	return client, nil
}
