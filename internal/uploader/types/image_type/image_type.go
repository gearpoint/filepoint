// image_type contains the image upload implementations.
package image_type

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/h2non/bimg"
)

const (
	// The uploader type key.
	Key types.UploaderTypes = "image"
	// defines the file quality.
	imageQuality = 60
)

const (
	// Defines the upload max size in bytes. Actual 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// Map of allowed images content types.
var ContentTypes = utils.ContentTypeMapping{
	"image/png":     "png",
	"image/jpeg":    "jpeg",
	"image/jpg":     "jpg",
	"image/svg+xml": "svg",
	"image/webp":    "webp",
	"image/tiff":    "tiff",
}

// NewUploaderType returns a new ImageUploader.
var NewUploaderType types.UploaderTypeLambda = func(uploaderTypeConfig *types.UploaderTypeConfig) types.UploaderType {
	return &ImageUploader{
		Config: uploaderTypeConfig,
	}
}

// ImageUploader is the image uploader implementation.
type ImageUploader struct {
	Config *types.UploaderTypeConfig
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

// GetConfig returns the uploader configuration
func (u *ImageUploader) GetConfig() *types.UploaderTypeConfig {
	return u.Config
}

// GetLabels returns the image labels.
func (u *ImageUploader) GetLabels(prefix string) []string {
	labels, _ := u.Config.AWSRepository.GetImageLabels(prefix)

	return labels
}

// HandleFile handles the image - converts it, etc.
func (u *ImageUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	file, err := u.Config.AWSRepository.DownloadFile(prefix)
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

	for contentType, ext := range ContentTypes {
		if ext == bimg.ImageTypes[imgConvert] {
			u.Config.UploadView.ContentType = contentType
			break
		}
	}

	return img, nil
}

// Upload uploads the image to S3.
func (u *ImageUploader) Upload(reader io.ReadCloser) (string, error) {
	cType := u.Config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.Config.UploadView.UserId,
		ContentTypes[cType],
	)

	var metadata = map[string]string{
		"user-id":  u.Config.UploadView.UserId,
		"title":    u.Config.UploadView.Title,
		"author":   u.Config.UploadView.Author,
		"filename": u.Config.UploadView.Filename,
	}

	err := u.Config.AWSRepository.UploadChunks(s3Prefix, reader, cType, metadata, nil)
	if err != nil {
		return "", err
	}

	return s3Prefix, nil
}

// UploadTemp uploads the image to S3 with lifecycle.
func (u *ImageUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	cType := u.Config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.Config.UploadView.UserId,
		ContentTypes[cType],
	)

	tagging := aws_repository.TempFileRule

	err := u.Config.AWSRepository.PutObject(s3Prefix, reader, cType, nil, &tagging)
	if err != nil {
		return "", errors.New("error uploading image to S3")
	}

	return s3Prefix, nil
}
