package views

import (
	"time"

	"github.com/gearpoint/filepoint/pkg/utils"
)

// UploadRequest contains the upload request body parameters.
type UploadRequest struct {
	UserId        string `form:"userId"`
	Title         string `form:"title"`
	Author        string `form:"author"`
	CorrelationId string `form:"correlationId"`
}

// GetSignedURLResponse is the response used in GetSignedURL calls.
type GetSignedURLResponse struct {
	Url       string            `json:"url"`
	Metadata  map[string]string `json:"metadata"`
	Tagging   map[string]string `json:"tagging"`
	Expires   time.Time         `json:"expires"`
	Temporary bool              `json:"temporary"`
}

// ListSignedURLResponse is the response for many GetSignedURLResponse fields
type ListSignedURLResponse map[string]*GetSignedURLResponse

// ListObjectsRequest is the request used in ListObjects calls.
type ListObjectsRequest struct {
	Prefixes   []string              `json:"prefixes"`
	Definition utils.FileDefinitions `json:"definition"`
}
