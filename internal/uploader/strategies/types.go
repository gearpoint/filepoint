// uploader is used to define uploader strategies (image, video, file...).
package strategies

import (
	"io"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// EventTypeKey is used to define the event type key of each strategy.
type EventTypeKey string

// FileDefinitions defines the available definitions.
type FileDefinitions string

// Uploader defines the file uploading methods.
type Uploader interface {
	Config() *UploaderConfig
	SetConfig(cfg *UploaderConfig)
	ContentTypes() utils.ContentTypeMapping
	SetContentTypes(contentTypes utils.ContentTypeMapping)
	FileDefinitions() utils.FileDefinitionsMapping
	SetFileDefinitions(fileDefinitions utils.FileDefinitionsMapping)
	Validate(uploadPubSub *views.UploadPubSub) error
	HandleFile(definition utils.FileDefinitions, tempFilename string) (io.ReadCloser, error)
	Upload(filename string, reader io.ReadCloser) (string, error)
	DownloadTemp(tempPrefix string) (io.ReadCloser, error)
	UploadTemp(reader io.ReadCloser) (string, error)
	SetLabels(filename string)
}

// UploaderConfig contains the uploader strategy configuration.
type UploaderConfig struct {
	UploadView    *views.UploadPubSub
	AWSRepository *aws_repository.AWSRepository
	Prefix        string
}
