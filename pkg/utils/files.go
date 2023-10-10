package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AleksK1NG/api-mc/pkg/httpErrors"
)

// defines the content type mapping structure.
type ContentTypeMapping map[string]string

// GetFileContentType returns the file content type.
func GetFileContentType(fileHeader textproto.MIMEHeader) (string, error) {
	contentTypes := fileHeader["Content-Type"]
	if len(contentTypes) < 1 {
		return "", httpErrors.NotAllowedImageHeader
	}

	return contentTypes[0], nil
}

// GetFileBytes parses a FileHeader instance to a byte array.
func GetFileBytes(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CreateTmpFile creates a new temporary file.
func CreateTmpFile(c *gin.Context, file []byte) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(file); err != nil {
		return "", err
	}

	return f.Name(), nil
}

// CheckAllowedContentType checks if the content file is allowed in the ContentTypeMapping.
func CheckAllowedContentType(allowedContentTypes ContentTypeMapping, contentType string) bool {
	_, allowed := allowedContentTypes[contentType]

	return allowed
}

// GetUniquePrefix returns an unique file prefix.
func GetUniquePrefix(userId string, ext string) string {
	randString := uuid.New().String()

	return userId + "/" + randString + "." + ext
}

// CheckPrefixIsFolder checks if the prefix is a folder.
func CheckPrefixIsFolder(prefix string) bool {
	split := strings.Split(prefix, "/")
	if len(split) > 1 {
		return false
	}

	return true
}

// GetPrefixFolder returns the prefix folder.
func GetPrefixFolder(prefix string) string {
	split := strings.Split(prefix, "/")
	if len(split) > 1 {
		return split[0]
	}

	return ""
}
