package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/AleksK1NG/api-mc/pkg/httpErrors"
)

// FileDefinitions is the type of the available file definitions.
type FileDefinitions int

const (
	LowDef FileDefinitions = iota
	MediumDef
	HighDef
)

// defines the file definitions prefix names.
type FileDefinitionsMapping map[FileDefinitions]string

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
func CreateTmpFile(reader io.ReadCloser) (string, error) {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}

// ReadCloserFromBytes returns a new ReadCloser instance from the given bytes
func ReadCloserFromBytes(b []byte) io.ReadCloser {
	newReader := bytes.NewReader(b)
	return io.NopCloser(newReader)
}

// CheckAllowedContentType checks if the content file is allowed in the ContentTypeMapping.
func CheckAllowedContentType(allowedContentTypes ContentTypeMapping, contentType string) bool {
	_, allowed := allowedContentTypes[contentType]

	return allowed
}

// CreatePrefix creates the prefix by joining the given strings.
// It returns a formatted prefix i.e a1/a2/a3.
func CreatePrefix(parts ...string) string {
	return strings.Join(parts, "/")
}

// GetUniquePrefix returns an unique folder prefix.
// It concatenates the given prefix string with a generated uuid.
func GetUniquePrefix(prefix string) string {
	randString := uuid.New().String()

	return CreatePrefix(prefix, randString)
}

// CheckPrefixIsFolder checks if the prefix is a folder.
// It verifies if the given prefix contains a file extension.
func CheckPrefixIsFolder(prefix string) bool {
	extension := filepath.Ext(prefix)

	return len(extension) == 0
}

// GetPrefixFolder returns the prefix folder and the sub folder depth.
// i.e:
//
// prefix := "my/prefix/test.txt"
// GetPrefixFolder(prefix) > "my/prefix", 2
func GetPrefixFolder(prefix string) (string, int) {
	split := strings.Split(
		strings.Trim(prefix, "/"), "/",
	)

	if len(split) == 1 {
		return "", 0
	}

	var prefixes []string
	for i := 0; i <= len(split)-2; i++ {
		prefixes = append(prefixes, split[i])
	}

	return CreatePrefix(prefixes...), len(prefixes)
}

// AtoFileDefinitions converts ascii to FileDefinitions.
func AtoFileDefinitions(def string) FileDefinitions {
	var definition FileDefinitions = MediumDef

	iDefinition, err := strconv.Atoi(def)
	if err == nil {
		definition = FileDefinitions(iDefinition)
	}

	return definition
}

// GetClosestPrefix returns the prefix with the closest definition (or exact if possible).
func GetClosestPrefix(definitionsMap FileDefinitionsMapping, definition FileDefinitions) string {
	keys := make([]int, 0, len(definitionsMap))

	for k := range definitionsMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	smallestKey := FileDefinitions(keys[0])
	if definition < smallestKey {
		return definitionsMap[smallestKey]
	}

	greaterKey := FileDefinitions(keys[len(keys)-1])
	if definition > greaterKey {
		return definitionsMap[greaterKey]
	}

	return definitionsMap[definition]
}
