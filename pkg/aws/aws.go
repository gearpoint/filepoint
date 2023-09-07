// package aws contains the Amazon Web Services functions and configurations.
package aws

import (
	"context"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gearpoint/filepoint/config"
)

// NewAWSClient returns a new AWS SDK v2 client instance.
func NewAWSClient(awsCfg *config.AWSConfig) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)

	return client, nil
}
