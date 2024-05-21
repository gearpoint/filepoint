package file_type

import (
	"context"
	"errors"
	"io"

	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// The uploader event type key.
	Key strategies.EventTypeKey = "file"
)

const (
	// Defines the upload max size in bytes. Actual: 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// FileUploader is the file uploader implementation.
type FileUploader struct {
	config       *strategies.UploaderConfig
	contentTypes utils.ContentTypeMapping
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	return &FileUploader{
		contentTypes: utils.ContentTypeMapping{
			"text/plain":      "txt",
			"application/pdf": "pdf",
		},
	}
}

// SetConfig adds the uploader configuration.
func (u *FileUploader) SetConfig(cfg *strategies.UploaderConfig) {
	u.config = cfg
}

// GetConfig returns the uploader configuration.
func (u *FileUploader) GetConfig() *strategies.UploaderConfig {
	return u.config
}

// GetContentTypes returns the uploader allowed content types.
func (u *FileUploader) GetContentTypes() utils.ContentTypeMapping {
	return u.contentTypes
}

// Validate validates the struct.
func (u *FileUploader) Validate(uploadPubSub *views.UploadPubSub) error {
	ctx := context.WithValue(
		context.Background(),
		utils.MaxFileSizeKey,
		uploadMaxSize,
	)

	return utils.Validate.StructCtx(ctx, uploadPubSub)
}

// SetLabels returns the file labels.
func (u *FileUploader) SetLabels(prefix string) {
}

// HandleFile handles the file - converts it, etc.
func (u *FileUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	return u.config.AWSRepository.DownloadFile(prefix)
}

// Upload uploads the file to S3.
func (u *FileUploader) Upload(reader io.ReadCloser) (string, error) {
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

// UploadTemp uploads the file to S3 with lifecycle.
func (u *FileUploader) UploadTemp(reader io.ReadCloser) (string, error) {
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
