// uploader is the upload director.
package uploader

import (
	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// Uploader is the upload director.
type Uploader struct {
	ContentTypes utils.ContentTypeMapping
	UploaderType types.UploaderType
}

// NewUploader returns a new Uploader instance.
func NewUploader(uploaderTypeMapping *types.TypeMapping, uploaderTypeConfig *types.UploaderTypeConfig) *Uploader {
	return &Uploader{
		ContentTypes: uploaderTypeMapping.ContentTypes,
		UploaderType: uploaderTypeMapping.NewUploaderType(uploaderTypeConfig),
	}
}
