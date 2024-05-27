package strategies

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// Defines the upload max size in bytes.
	uploadMaxSize int64 = 0
)

// BaseUploader is the Uploader base implementation.
// This must be implemented by a file type strategy.
type BaseUploader struct {
	config          *UploaderConfig
	contentTypes    utils.ContentTypeMapping
	fileDefinitions utils.FileDefinitionsMapping
}

// SetConfig adds the uploader configuration.
func (u *BaseUploader) SetConfig(cfg *UploaderConfig) {
	u.config = cfg
}

// GetConfig returns the uploader configuration.
func (u *BaseUploader) GetConfig() *UploaderConfig {
	return u.config
}

// GetContentTypes returns the uploader allowed content types.
func (u *BaseUploader) GetContentTypes() utils.ContentTypeMapping {
	return u.contentTypes
}

// FormatPrefix formats the file prefix with the given filename.
// It appends the folder prefix and uses the configured content type as extension.
func (u *BaseUploader) FormatPrefix(filename string) string {
	filename = fmt.Sprintf("%s.%s", filename, u.contentTypes[u.config.UploadView.ContentType])

	return utils.CreatePrefix(u.config.Prefix, filename)
}

// GetFileDefinitions returns the file definitions.
func (u *BaseUploader) GetFileDefinitions() utils.FileDefinitionsMapping {
	return u.fileDefinitions
}

// Validate validates the struct.
func (u *BaseUploader) Validate(uploadPubSub *views.UploadPubSub) error {
	ctx := context.WithValue(
		context.Background(),
		utils.MaxFileSizeKey,
		uploadMaxSize,
	)

	return utils.Validate.StructCtx(ctx, uploadPubSub)
}

// HandleFile handles the video - converts it, etc.
func (u *BaseUploader) HandleFile(definition utils.FileDefinitions, reader io.ReadCloser) (io.ReadCloser, error) {
	return reader, nil
}

// Upload uploads a new file to S3.
func (u *BaseUploader) Upload(filename string, reader io.ReadCloser) (string, error) {
	s3Prefix := u.FormatPrefix(filename)

	var metadata = map[string]string{
		"user-id":  u.config.UploadView.UserId,
		"title":    u.config.UploadView.Title,
		"author":   u.config.UploadView.Author,
		"filename": u.config.UploadView.Filename,
	}

	err := u.config.AWSRepository.PutObject(
		s3Prefix,
		reader,
		u.config.UploadView.ContentType,
		metadata,
		nil,
	)
	if err != nil {
		return "", err
	}

	return s3Prefix, nil
}

// UploadTemp uploads the file to S3 with lifecycle.
func (u *BaseUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	s3Prefix := u.FormatPrefix(aws_repository.TempFileRule)
	tagging := aws_repository.TempFileRule

	err := u.config.AWSRepository.PutObject(
		s3Prefix,
		reader,
		u.config.UploadView.ContentType,
		nil,
		&tagging,
	)
	if err != nil {
		return "", errors.New("error uploading file to S3")
	}

	return s3Prefix, nil
}

// DownloadTemp downloads the temp file from S3.
func (u *BaseUploader) DownloadTemp(tempPrefix string) (io.ReadCloser, error) {
	return u.config.AWSRepository.DownloadFile(tempPrefix)
}

// SetLabels starts the file label detection.
func (u *BaseUploader) SetLabels(filename string) {
}
