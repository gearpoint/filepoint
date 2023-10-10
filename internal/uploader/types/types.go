// types is used to control the upload types (image, video, file...).
package types

import (
	"io"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// Uploader types definition.
type UploaderTypes string

// UploaderTypeLambda is used to return the uploader type.
type UploaderTypeLambda func(*UploaderTypeConfig) UploaderType

// UploaderType contains the methods of any type of upload.
type UploaderType interface {
	Validate(uploadPubSub *views.UploadPubSub) error
	GetConfig() *UploaderTypeConfig
	GetLabels(prefix string) []string
	HandleFile(prefix string) (io.ReadCloser, error)
	Upload(reader io.ReadCloser) (string, error)
	UploadTemp(reader io.ReadCloser) (string, error)
}

// UploaderTypeConfig contains the uploader types configuration.
type UploaderTypeConfig struct {
	UploadView    *views.UploadPubSub
	AWSRepository *aws_repository.AWSRepository
}

// TypeMapping defines each implemented type.
type TypeMapping struct {
	NewUploaderType UploaderTypeLambda
	ContentTypes    utils.ContentTypeMapping
}
