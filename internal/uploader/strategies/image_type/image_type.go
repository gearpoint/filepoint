// image_type contains the image upload implementations.
package image_type

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/h2non/bimg"
)

const (
	// The uploader event type key.
	Key strategies.EventTypeKey = "image"
	// defines the file quality.
	imageQuality = 60
)

const (
	// Defines the upload max size in bytes. Actual: 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// ImageUploader is the image uploader implementation.
type ImageUploader struct {
	config       *strategies.UploaderConfig
	contentTypes utils.ContentTypeMapping
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	return &ImageUploader{
		contentTypes: utils.ContentTypeMapping{
			"image/png":     "png",
			"image/jpeg":    "jpeg",
			"image/jpg":     "jpg",
			"image/svg+xml": "svg",
			"image/webp":    "webp",
			"image/tiff":    "tiff",
		},
	}
}

// SetConfig adds the uploader configuration.
func (u *ImageUploader) SetConfig(cfg *strategies.UploaderConfig) {
	u.config = cfg
}

// GetConfig returns the uploader configuration.
func (u *ImageUploader) GetConfig() *strategies.UploaderConfig {
	return u.config
}

// GetContentTypes returns the uploader allowed content types.
func (u *ImageUploader) GetContentTypes() utils.ContentTypeMapping {
	return u.contentTypes
}

// Validate validates the struct.
func (u *ImageUploader) Validate(uploadPubSub *views.UploadPubSub) error {
	ctx := context.WithValue(
		context.Background(),
		utils.MaxFileSizeKey,
		uploadMaxSize,
	)

	return utils.Validate.StructCtx(ctx, uploadPubSub)
}

// GetLabels returns the image labels.
func (u *ImageUploader) SetLabels(prefix string) {
	u.config.AWSRepository.GetImageLabels(prefix)
	// todo: save labels in DynamoDB
}

// HandleFile handles the image - converts it, etc.
func (u *ImageUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	file, err := u.config.AWSRepository.DownloadFile(prefix)
	if err != nil {
		return nil, err
	}

	buffer, err := io.ReadAll(file)
	file.Close()

	if err != nil {
		return nil, err
	}

	image, err := u.handleImage(buffer)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(image)
	readCloser := io.NopCloser(reader)

	return readCloser, nil
}

// handleImage proccess the image.
func (u *ImageUploader) handleImage(buffer []byte) ([]byte, error) {
	imgConvert := bimg.JPEG
	if bimg.DetermineImageType(buffer) == imgConvert {
		imgConvert = 0
	}

	img, err := bimg.NewImage(buffer).Process(
		bimg.Options{
			Quality: 60,
			Type:    imgConvert,
			Speed:   8,
		},
	)

	if err != nil {
		return nil, err
	}

	for contentType, ext := range u.contentTypes {
		if ext == bimg.ImageTypes[imgConvert] {
			u.config.UploadView.ContentType = contentType
			break
		}
	}

	return img, nil
}

// Upload uploads the image to S3.
func (u *ImageUploader) Upload(reader io.ReadCloser) (string, error) {
	cType := u.config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.config.UploadView.UserId,
		u.contentTypes[cType],
	)

	var metadata = map[string]string{
		"user-id":  u.config.UploadView.UserId,
		"title":    u.config.UploadView.Title,
		"author":   u.config.UploadView.Author,
		"filename": u.config.UploadView.Filename,
	}

	err := u.config.AWSRepository.UploadChunks(s3Prefix, reader, cType, metadata, nil)
	if err != nil {
		return "", err
	}

	return s3Prefix, nil
}

// UploadTemp uploads the image to S3 with lifecycle.
func (u *ImageUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	cType := u.config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.config.UploadView.UserId,
		u.contentTypes[cType],
	)

	tagging := aws_repository.TempFileRule

	err := u.config.AWSRepository.PutObject(s3Prefix, reader, cType, nil, &tagging)
	if err != nil {
		return "", errors.New("error uploading image to S3")
	}

	return s3Prefix, nil
}
