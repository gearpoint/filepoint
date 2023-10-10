package uploader

import (
	"errors"

	"github.com/gearpoint/filepoint/internal/uploader/types"
	"github.com/gearpoint/filepoint/internal/uploader/types/file_type"
	"github.com/gearpoint/filepoint/internal/uploader/types/image_type"
	"github.com/gearpoint/filepoint/internal/uploader/types/video_type"
	"github.com/gearpoint/filepoint/pkg/utils"
)

// The available types mapping.
var UploaderTypesMap = map[types.UploaderTypes]types.TypeMapping{
	file_type.Key: {
		NewUploaderType: file_type.NewUploaderType,
		ContentTypes:    file_type.ContentTypes,
	},
	image_type.Key: {
		NewUploaderType: image_type.NewUploaderType,
		ContentTypes:    image_type.ContentTypes,
	},
	video_type.Key: {
		NewUploaderType: video_type.NewUploaderType,
		ContentTypes:    video_type.ContentTypes,
	},
}

// GetTypes returns the uploader types
func GetTypes() []types.UploaderTypes {
	var types []types.UploaderTypes
	for key := range UploaderTypesMap {
		types = append(types, key)
	}

	return types
}

// GetTypeMappingByKey returns the uploader type mapping by the key.
func GetTypeMappingByKey(key types.UploaderTypes) (*types.TypeMapping, error) {
	for typ, uploaderType := range UploaderTypesMap {
		if key == typ {
			return &uploaderType, nil
		}
	}

	return nil, errors.New("uploader type mapping not found")
}

// GetTypeByContentType checks if the content type is present in any content type mapping.
func GetTypeByContentType(contentType string) (types.UploaderTypes, *types.TypeMapping, error) {
	for key, uploaderType := range UploaderTypesMap {
		if utils.CheckAllowedContentType(uploaderType.ContentTypes, contentType) {
			return key, &uploaderType, nil
		}
	}

	return "", nil, errors.New("invalid content type")
}
