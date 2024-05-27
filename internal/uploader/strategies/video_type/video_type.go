package video_type

import (
	"errors"
	"io"

	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// The uploader event type key.
	Key strategies.EventTypeKey = "video"

	// Defines the upload max size in bytes. Current: 1 gibibyte.
	uploadMaxSize int64 = 1 << 30
)

// VideoUploader is the video uploader implementation.
type VideoUploader struct {
	strategies.BaseUploader

	config          *strategies.UploaderConfig
	contentTypes    utils.ContentTypeMapping
	fileDefinitions utils.FileDefinitionsMapping
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	return &VideoUploader{
		contentTypes: utils.ContentTypeMapping{
			"video/mp4":  "mp4",
			"video/mpeg": "mpeg",
			"video/ogg":  "ogv",
		},
		fileDefinitions: utils.FileDefinitionsMapping{
			utils.HighDef: "high-def",
		},
	}
}

// HandleFile handles the video - converts it, etc.
func (u *VideoUploader) HandleFile(definition utils.FileDefinitions, reader io.ReadCloser) (io.ReadCloser, error) {
	// todo: add goffmpeg
	// do not use ReadAll with large files
	return reader, nil
}

// Upload uploads a new file to S3.
func (u *VideoUploader) Upload(filename string, reader io.ReadCloser) (string, error) {
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
func (u *VideoUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	s3Prefix := u.FormatPrefix(aws_repository.TempFileRule)
	tagging := aws_repository.TempFileRule

	err := u.config.AWSRepository.UploadChunks(
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

// SetLabels starts the video label detection.
func (u *VideoUploader) SetLabels(filename string) {
	// todo: save labels in DynamoDB
	// s3Prefix := u.FormatPrefix(filename)
	// u.config.AWSRepository.StartVideoLabelsDetection(s3Prefix)
}
