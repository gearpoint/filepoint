package file_type

import (
	"context"
	"errors"
	"io"

	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// The uploader type key.
	Key types.UploaderTypes = "file"
)

const (
	// Defines the upload max size in bytes. Actual 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// Map of allowed files content types.
var ContentTypes = utils.ContentTypeMapping{
	"text/plain":      "txt",
	"application/pdf": "pdf",
}

// NewUploaderType returns a new FileUploader.
var NewUploaderType types.UploaderTypeLambda = func(uploaderTypeConfig *types.UploaderTypeConfig) types.UploaderType {
	return &FileUploader{
		Config: uploaderTypeConfig,
	}
}

// FileUploader is the file uploader implementation.
type FileUploader struct {
	Config *types.UploaderTypeConfig
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

// GetConfig returns the uploader configuration
func (u *FileUploader) GetConfig() *types.UploaderTypeConfig {
	return u.Config
}

// GetLabels returns the file labels.
func (u *FileUploader) GetLabels(prefix string) []string {
	return nil
}

// HandleFile handles the file - converts it, etc.
func (u *FileUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	return u.Config.AWSRepository.DownloadFile(prefix)
}

// Upload uploads the file to S3.
func (u *FileUploader) Upload(reader io.ReadCloser) (string, error) {
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

// UploadTemp uploads the file to S3 with lifecycle.
func (u *FileUploader) UploadTemp(reader io.ReadCloser) (string, error) {
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
