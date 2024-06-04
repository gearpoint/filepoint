// image_type contains the image upload implementations.
package image_type

import (
	"io"
	"os"

	"github.com/gearpoint/filepoint/internal/uploader/strategies"
	"github.com/gearpoint/filepoint/pkg/utils"
	"github.com/h2non/bimg"
)

const (
	// The uploader event type key.
	Key strategies.EventTypeKey = "image"

	// Defines the upload max size in bytes. Current: 15 mebibytes.
	uploadMaxSize int64 = 15 << 20
)

// ImageUploader is the image uploader implementation.
type ImageUploader struct {
	strategies.BaseUploader
}

// NewUploader returns a new Uploader instance.
func NewUploader() strategies.Uploader {
	uploader := &ImageUploader{
		BaseUploader: strategies.BaseUploader{
			UploadMaxSize: uploadMaxSize,
		},
	}
	uploader.SetContentTypes(utils.ContentTypeMapping{
		"image/png":     "png",
		"image/jpeg":    "jpeg",
		"image/jpg":     "jpg",
		"image/svg+xml": "svg",
		"image/webp":    "webp",
		"image/tiff":    "tiff",
	})
	uploader.SetFileDefinitions(utils.FileDefinitionsMapping{
		utils.LowDef:    "low-def",
		utils.MediumDef: "medium-def",
		utils.HighDef:   "high-def",
	})

	return uploader
}

// HandleFile handles the image - converts it, etc.
func (u *ImageUploader) HandleFile(definition utils.FileDefinitions, tempFilename string) (io.ReadCloser, error) {
	buffer, err := os.ReadFile(tempFilename)
	if err != nil {
		return nil, err
	}

	image, err := u.handleImage(buffer, definition)
	if err != nil {
		return nil, err
	}

	return utils.ReadCloserFromBytes(image), nil
}

// handleImage proccess the image.
func (u *ImageUploader) handleImage(buffer []byte, definition utils.FileDefinitions) ([]byte, error) {
	processingOpts := u.getProccessingOptions(definition)
	img, err := bimg.NewImage(buffer).Process(
		processingOpts,
	)

	if err != nil {
		return nil, err
	}

	// Changes the current ContentType configured in the instance.
	for contentType, ext := range u.ContentTypes() {
		if ext == bimg.ImageTypes[processingOpts.Type] {
			u.Config().UploadView.ContentType = contentType
			break
		}
	}

	return img, nil
}

// getProccessingOptions returns the bimg options according to the image definition.
func (u *ImageUploader) getProccessingOptions(definition utils.FileDefinitions) bimg.Options {
	options := bimg.Options{
		Type:         bimg.WEBP,
		Speed:        7,
		Embed:        true,
		NoAutoRotate: true,
		Force:        false,
		Enlarge:      true,
		Quality:      85,
	}

	type processingOptions func() bimg.Options

	ruleset := map[utils.FileDefinitions]processingOptions{
		utils.LowDef: func() bimg.Options {
			options.Height = 360
			options.Compression = 14
			return options
		},
		utils.MediumDef: func() bimg.Options {
			options.Height = 720
			options.Compression = 10
			return options
		},
		utils.HighDef: func() bimg.Options {
			options.Height = 1920
			return options
		},
	}

	return ruleset[definition]()
}

// SetLabels starts the image image rekognition labels.
func (u *ImageUploader) SetLabels(filename string) {
	// todo: save labels in DynamoDB
	// todo: make sure it's jpeg or png
	// s3Prefix := u.FormatPrefix(filename)
	// u.Config().AWSRepository.GetImageLabels(s3Prefix)
}
