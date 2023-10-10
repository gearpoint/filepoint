package video_type

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
	Key types.UploaderTypes = "video"
)

const (
	// Defines the upload max size in bytes. Actual 1 gibibyte.
	uploadMaxSize int64 = 1 << 30
)

// Map of allowed videos content types.
var ContentTypes = utils.ContentTypeMapping{
	"video/mp4":  "mp4",
	"video/mpeg": "mpeg",
	"video/ogg":  "ogv",
}

// NewUploaderType returns a new VideoUploader.
var NewUploaderType types.UploaderTypeLambda = func(uploaderTypeConfig *types.UploaderTypeConfig) types.UploaderType {
	return &VideoUploader{
		Config: uploaderTypeConfig,
	}
}

// VideoUploader is the video uploader implementation.
type VideoUploader struct {
	Config *types.UploaderTypeConfig
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

// GetConfig returns the uploader configuration
func (u *VideoUploader) GetConfig() *types.UploaderTypeConfig {
	return u.Config
}

// GetLabels starts the video label detection.
func (u *VideoUploader) GetLabels(prefix string) []string {
	labels, _ := u.Config.AWSRepository.StartVideoLabelsDetection(prefix)

	return labels
}

// HandleFile handles the video - converts it, etc.
func (u *VideoUploader) HandleFile(prefix string) (io.ReadCloser, error) {
	return u.Config.AWSRepository.DownloadFile(prefix)
	// goffmpeg?
}

func (u *VideoUploader) Upload(reader io.ReadCloser) (string, error) {
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

// UploadTemp uploads the video to S3 with lifecycle.
func (u *VideoUploader) UploadTemp(reader io.ReadCloser) (string, error) {
	cType := u.Config.UploadView.ContentType
	s3Prefix := utils.GetUniquePrefix(
		u.Config.UploadView.UserId,
		ContentTypes[cType],
	)

	tagging := aws_repository.TempFileRule

	err := u.Config.AWSRepository.UploadChunks(s3Prefix, reader, cType, nil, &tagging)
	if err != nil {
		return "", errors.New("error uploading video to S3")
	}

	return s3Prefix, nil
}
