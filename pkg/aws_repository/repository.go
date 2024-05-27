package aws_repository

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// Temporary file rule. This rule must be configured at the defined bucket.
	TempFileRule = "temporary-file"
	// Signed url expiration time.
	SignExpiration = 12 * time.Hour
)

// The Rekognition unsuported locations.
var unsupportedRekoLocations = map[string]bool{
	"sa-east-1": true,
}

// AWSRepository contains the aws config and implementations.
type AWSRepository struct {
	ctx                  context.Context
	config               *cfg.AWSConfig
	s3Client             *s3.Client
	rekoClient           *rekognition.Client
	dynamoClient         *dynamodb.Client
	cloudfrontDist       string
	cloudfrontPrivateKey rsa.PrivateKey
}

// NewAWSRepository returns a AWSRepository instance.
func NewAWSRepository(awsConfig *cfg.AWSConfig, ctx context.Context) (*AWSRepository, error) {
	sdkConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsConfig.Region),
		getEndpointResolver(awsConfig),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.UsePathStyle = utils.IsDevEnvironment()
	})

	if unsupportedRekoLocations[sdkConfig.Region] {
		sdkConfig.Region = "us-east-1"
	}
	rekoClient := rekognition.NewFromConfig(sdkConfig)

	dynamoClient := dynamodb.NewFromConfig(sdkConfig)

	rsaKey, err := getPrivateKey(awsConfig.CloudfrontCrtFile)
	if err != nil {
		return nil, err
	}

	return &AWSRepository{
		ctx:                  ctx,
		config:               awsConfig,
		s3Client:             s3Client,
		rekoClient:           rekoClient,
		dynamoClient:         dynamoClient,
		cloudfrontDist:       awsConfig.CloudfrontDist,
		cloudfrontPrivateKey: *rsaKey,
	}, nil
}

// getEndpointResolver configures the AWS endpoint that will be used to make API calls.
func getEndpointResolver(awsConfig *cfg.AWSConfig) config.LoadOptionsFunc {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsConfig.Endpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsConfig.Endpoint,
				SigningRegion: awsConfig.Region,
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to its default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	return config.WithEndpointResolverWithOptions(customResolver)
}

// getPrivateKey returns a private key from a file path.
func getPrivateKey(filepath string) (*rsa.PrivateKey, error) {
	return sign.LoadPEMPrivKeyFile(filepath)
}

// CheckIsNotFoundError checks if the aws error is not found.
func CheckIsNotFoundError(err error) bool {
	var responseError *awshttp.ResponseError
	if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
		return true
	}

	return false
}
