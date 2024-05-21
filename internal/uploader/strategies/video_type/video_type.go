package video_type

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
	Key strategies.EventTypeKey = "video"
)

const (
	// Defines the upload max size in bytes. Actual: 1 gibibyte.
	uploadMaxSize int64 = 1 << 30
)

// VideoUploader is the video uploader implementation.
type VideoUploader struct {
	config       *strategies.UploaderConfig
	contentTypes utils.ContentTypeMapping
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	return &VideoUploader{
		contentTypes: utils.ContentTypeMapping{
			"video/mp4":  "mp4",
			"video/mpeg": "mpeg",
			"video/ogg":  "ogv",
		},
	}
}

// SetConfig adds the uploader configuration.
func (u *VideoUploader) SetConfig(cfg *strategies.UploaderConfig) {
	u.config = cfg
}

// GetConfig returns the uploader configuration.
func (u *VideoUploader) GetConfig() *strategies.UploaderConfig {
	return u.config
}

// GetContentTypes returns the uploader allowed content types.
func (u *VideoUploader) GetContentTypes() utils.ContentTypeMapping {
	return u.contentTypes
}

// Validate validates the struct.
func (u *VideoUploader) Validate(uploadPubSub *views.UploadPubSub) error {
	ctx := context.WithValue(
		context.Background(),
		utils.MaxFileSizeKey,
		uploadMaxSize,
	)

	return utils.Validate.StructCtx(ctx, uploadPubSub)
}

// SetLabels starts the video label detection.
func (u *VideoUploader) SetLabels(prefix string) {
	u.config.AWSRepository.StartVideoLabelsDetection(prefix)
}

// HandleFile handles the video - converts it, etc.
func (u *VideoUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	// todo: goffmpeg
	return u.config.AWSRepository.DownloadFile(prefix)
}

func (u *VideoUploader) Upload(reader io.ReadCloser) (string, error) {
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

// UploadTemp uploads the video to S3 with lifecycle.
func (u *VideoUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	cType := u.config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.config.UploadView.UserId,
		u.contentTypes[cType],
	)

	tagging := aws_repository.TempFileRule

	err := u.config.AWSRepository.UploadChunks(s3Prefix, reader, cType, nil, &tagging)
	if err != nil {
		return "", errors.New("error uploading video to S3")
	}

	return s3Prefix, nil
}
