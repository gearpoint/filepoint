package aws_repository

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	cfg "github.com/gearpoint/filepoint/config"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/logger"
	"go.uber.org/zap"
)

const (
	// Temporary file rule.
	TempFileRule = "temporary-file"
	// Signed url expiration time.
	SignExpiration = 12 * time.Hour
	// The max number of labels returned.
	maxRekognitionLabels int32 = 10
	// Cloudfront private key file location.
	cloudfrontPrivateKey = ".aws/filepoint-cloudfront.pem"
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
	cloudfrontDist       string
	cloudfrontPrivateKey rsa.PrivateKey
}

// NewAWSRepository returns a AWSRepository instance.
func NewAWSRepository(awsConfig *cfg.AWSConfig, ctx context.Context) (*AWSRepository, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsConfig.Region))
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(sdkConfig)

	if unsupportedRekoLocations[sdkConfig.Region] {
		sdkConfig.Region = "us-east-1"
	}
	rekoClient := rekognition.NewFromConfig(sdkConfig)

	rsaKey, err := getPrivateKey(cloudfrontPrivateKey)
	if err != nil {
		return nil, err
	}

	return &AWSRepository{
		ctx:                  ctx,
		config:               awsConfig,
		s3Client:             s3Client,
		rekoClient:           rekoClient,
		cloudfrontDist:       awsConfig.CloudfrontDist,
		cloudfrontPrivateKey: *rsaKey,
	}, nil
}

// PutObject puts a new object in the given prefix.
// Do not use it for large files.
func (r *AWSRepository) PutObject(prefix string, file io.Reader, contentType string, metadata map[string]string, tagging *string) error {
	_, err := r.s3Client.PutObject(r.ctx, &s3.PutObjectInput{
		Bucket:      &r.config.Bucket,
		Key:         &prefix,
		Body:        file,
		ContentType: &contentType,
		Metadata:    metadata,
		Tagging:     tagging,
	})

	return err
}

// UploadChunks puts a new object in the given prefix.
func (r *AWSRepository) UploadChunks(prefix string, file io.Reader, contentType string, metadata map[string]string, tagging *string) error {
	uploader := manager.NewUploader(r.s3Client)
	_, err := uploader.Upload(r.ctx, &s3.PutObjectInput{
		Bucket:      &r.config.Bucket,
		Key:         &prefix,
		Body:        file,
		ContentType: &contentType,
		Metadata:    metadata,
		Tagging:     tagging,
	})

	return err
}

// DownloadFile gets an object from a bucket and returns it.
func (r *AWSRepository) DownloadFile(prefix string) (io.ReadCloser, error) {
	result, err := r.s3Client.GetObject(r.ctx, &s3.GetObjectInput{
		Bucket: &r.config.Bucket,
		Key:    &prefix,
	})
	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

// GetSignedObject returns a Signed object from the given prefix.
func (r *AWSRepository) GetSignedObject(prefix string) (*views.GetSignedURLResponse, error) {
	defaultErr := errors.New("an internal error occured")

	obj, err := r.headObject(prefix)
	if err != nil {
		return nil, err
	}

	labels, temp, err := r.GetObjectTagging(prefix)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", r.cloudfrontDist, prefix)
	expires := time.Now().Add(SignExpiration)

	signer := sign.NewURLSigner(r.config.CloudfrontKeyId, &r.cloudfrontPrivateKey)
	signedUrl, err := signer.Sign(url, expires)
	if err != nil {
		return nil, defaultErr
	}

	return &views.GetSignedURLResponse{
		Url:       signedUrl,
		Metadata:  obj.Metadata,
		Labels:    labels,
		Expires:   expires,
		Temporary: temp,
	}, nil
}

// ListObjects lists all objects in the given prefix.
func (r *AWSRepository) ListObjects(prefix string) ([]string, error) {
	defaultErr := errors.New("an internal error occured")

	var mu sync.Mutex
	var wg sync.WaitGroup

	response := []string{}

	params := &s3.ListObjectsV2Input{
		Bucket: &r.config.Bucket,
		Prefix: &prefix,
	}
	paginator := s3.NewListObjectsV2Paginator(r.s3Client, params)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(r.ctx)
		if err != nil {
			logger.Error("error listing objects", zap.Error(err))

			return nil, defaultErr
		}

		for _, obj := range page.Contents {
			wg.Add(1)
			go func(obj s3types.Object) {
				defer wg.Done()
				mu.Lock()
				response = append(response, *obj.Key)
				mu.Unlock()
			}(obj)
		}
	}
	wg.Wait()

	return response, nil
}

// DeleteObject deletes an object from the given prefix.
func (r *AWSRepository) DeleteObject(prefix string) error {
	_, err := r.s3Client.DeleteObject(r.ctx, &s3.DeleteObjectInput{
		Bucket: &r.config.Bucket,
		Key:    &prefix,
	})

	return err
}

// DeleteMany deletes many objects.
func (r *AWSRepository) DeleteMany(prefixes []string) error {
	var objects []s3types.ObjectIdentifier
	for _, prefix := range prefixes {
		p := prefix
		objects = append(objects, s3types.ObjectIdentifier{
			Key: &p,
		})
	}
	_, err := r.s3Client.DeleteObjects(r.ctx, &s3.DeleteObjectsInput{
		Bucket: &r.config.Bucket,
		Delete: &s3types.Delete{
			Objects: objects,
		},
	})

	return err
}

// headObject returns the object infos.
func (r *AWSRepository) headObject(prefix string) (*s3.HeadObjectOutput, error) {
	return r.s3Client.HeadObject(r.ctx, &s3.HeadObjectInput{
		Bucket: &r.config.Bucket,
		Key:    &prefix,
	})
}

// PutObjectTagging adds tags in the object.
func (r *AWSRepository) PutObjectTagging(prefix string, tagging map[string]string) error {
	tags := make([]s3types.Tag, len(tagging))

	i := 0
	for tagKey, tagValue := range tagging {
		tags[i] = s3types.Tag{
			Key:   aws.String(tagKey),
			Value: aws.String(tagValue),
		}
		i++
	}

	_, err := r.s3Client.PutObjectTagging(r.ctx, &s3.PutObjectTaggingInput{
		Bucket: &r.config.Bucket,
		Key:    &prefix,
		Tagging: &s3types.Tagging{
			TagSet: tags,
		},
	})
	if err != nil {
		return errors.New("error setting object tags")
	}

	return nil
}

// GetObjectTagging gets the object tags.
func (r *AWSRepository) GetObjectTagging(prefix string) ([]string, bool, error) {
	response, err := r.s3Client.GetObjectTagging(r.ctx, &s3.GetObjectTaggingInput{
		Bucket: &r.config.Bucket,
		Key:    &prefix,
	})
	if err != nil {
		return nil, false, errors.New("error getting object tags")
	}

	var temporary bool
	tags := []string{}
	for _, tag := range response.TagSet {
		if *tag.Key == TempFileRule {
			temporary = true
		}

		tags = append(tags, *tag.Key)
	}
	return tags, temporary, nil
}

// getPrivateKey returns a private key.
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
