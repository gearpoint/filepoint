package file_type

import (
	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/pkg/utils"
)

const (
	// The uploader event type key.
	Key strategies.EventTypeKey = "file"

	// Defines the upload max size in bytes. Current: 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// FileUploader is the file uploader implementation.
type FileUploader struct {
	strategies.BaseUploader

	config          *strategies.UploaderConfig
	contentTypes    utils.ContentTypeMapping
	fileDefinitions utils.FileDefinitionsMapping
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	return &FileUploader{
		contentTypes: utils.ContentTypeMapping{
			"text/plain":      "txt",
			"application/pdf": "pdf",
		},
		fileDefinitions: utils.FileDefinitionsMapping{
			utils.HighDef: "high-def",
		},
	}
}
