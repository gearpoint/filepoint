// strategies is used to control the uploader strategies (image, video, file...).
package strategies

import (
	"io"

	"github.com/gearpoint/filepoint/internal/views"
	"github.com/gearpoint/filepoint/pkg/aws_repository"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// EventTypeKey is used to define the event type key of each strategy.
type EventTypeKey string

// Uploader contains the methods of any uploader.
type Uploader interface {
	SetConfig(cfg *UploaderConfig)
	GetConfig() *UploaderConfig
	GetContentTypes() utils.ContentTypeMapping
	Validate(uploadPubSub *views.UploadPubSub) error
	SetLabels(prefix string)
	HandleFile(prefix string) (io.ReadCloser, error)
	Upload(reader io.ReadCloser) (string, error)
	UploadTemp(reader io.ReadCloser) (string, error)
}

// UploaderConfig contains the uploader strategy configuration.
type UploaderConfig struct {
	UploadView    *views.UploadPubSub
	AWSRepository *aws_repository.AWSRepository
}
