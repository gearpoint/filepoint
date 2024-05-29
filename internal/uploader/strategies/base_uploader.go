package strategies

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// BaseUploader is the Uploader base implementation.
// This must be implemented by a file type strategy.
type BaseUploader struct {
	config          *UploaderConfig
	contentTypes    utils.ContentTypeMapping
	fileDefinitions utils.FileDefinitionsMapping

	// Defines the upload max size in bytes.
	UploadMaxSize int64
}

// Config returns the uploader configuration.
func (u *BaseUploader) Config() *UploaderConfig {
	return u.config
}

// SetConfig adds the uploader configuration.
func (u *BaseUploader) SetConfig(cfg *UploaderConfig) {
	u.config = cfg
}

// ContentTypes returns the uploader allowed content types.
func (u *BaseUploader) ContentTypes() utils.ContentTypeMapping {
	return u.contentTypes
}

// SetContentTypes adds the allowed content types.
func (u *BaseUploader) SetContentTypes(contentTypes utils.ContentTypeMapping) {
	u.contentTypes = contentTypes
}

// FileDefinitions returns the file definitions.
func (u *BaseUploader) FileDefinitions() utils.FileDefinitionsMapping {
	return u.fileDefinitions
}

// SetFileDefinitions adds the allowed file definitions.
func (u *BaseUploader) SetFileDefinitions(fileDefinitions utils.FileDefinitionsMapping) {
	u.fileDefinitions = fileDefinitions
}

// FormatPrefix formats the file prefix with the given filename.
// It appends the folder prefix and uses the configured content type as extension.
func (u *BaseUploader) FormatPrefix(filename string) string {
	filename = fmt.Sprintf("%s.%s", filename, u.contentTypes[u.config.UploadView.ContentType])

	return utils.CreatePrefix(u.config.Prefix, filename)
}

// SetFileDefinitions returns the file definitions.
func (u *BaseUploader) GetFileDefinitions() utils.FileDefinitionsMapping {
	return u.fileDefinitions
}

// Validate validates the struct.
func (u *BaseUploader) Validate(uploadPubSub *views.UploadPubSub) error {
	ctx := context.WithValue(
		context.Background(),
		utils.MaxFileSizeKey,
		u.UploadMaxSize,
	)

	return utils.Validate.StructCtx(ctx, uploadPubSub)
}

// HandleFile handles the video - converts it, etc.
func (u *BaseUploader) HandleFile(definition utils.FileDefinitions, tempFilename string) (io.ReadCloser, error) {
	file, err := os.Open(tempFilename)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(file), nil
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

	err := u.config.AWSRepository.UploadChunks(
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
