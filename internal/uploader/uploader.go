package uploader

import (
	"errors"

	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/internal/uploader/strategies/file_type"
	"github.com/gearpoint/filepoint/internal/uploader/strategies/image_type"
	"github.com/gearpoint/filepoint/internal/uploader/strategies/video_type"
	"github.com/gearpoint/filepoint/pkg/utils"
)

type initUploader func() strategies.Uploader

var uploadersMap = map[strategies.EventTypeKey]initUploader{
	file_type.Key:  file_type.NewUploader,
	image_type.Key: image_type.NewUploader,
	video_type.Key: video_type.NewUploader,
}

// GetUploaderByEventType returns the uploader type mapping by the event type.
func GetUploaderByEventType(event_type strategies.EventTypeKey) (strategies.Uploader, error) {
	for key, initUploader := range uploadersMap {
		if key == event_type {
			return initUploader(), nil
		}
	}

	return nil, errors.New("uploader type mapping not found")
}

// GetUploaderByContentType checks if the content type is present in any content type mapping.
// It then returns the strategy related to the content type.
func GetUploaderByContentType(contentType string) (strategies.EventTypeKey, strategies.Uploader, error) {
	for key, initUploader := range uploadersMap {
		uploader := initUploader()
		if utils.CheckAllowedContentType(uploader.GetContentTypes(), contentType) {
			return key, uploader, nil
		}
	}

	return "", nil, errors.New("invalid content type")
}
