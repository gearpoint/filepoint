package utils

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
)

var (
	// ErrInvalidFileSignature is returned when file magic bytes don't match content type
	ErrInvalidFileSignature = errors.New("file signature does not match content type")

	// ErrUnsupportedFileType is returned when file type is not supported
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

// FileSignature represents a file's magic byte signature
type FileSignature struct {
	ContentType string
	Signatures  [][]byte
	Offset      int // offset where signature starts
}

// Common file signatures (magic bytes)
var fileSignatures = []FileSignature{
	// Image formats
	{
		ContentType: "image/png",
		Signatures: [][]byte{
			{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG signature
		},
		Offset: 0,
	},
	{
		ContentType: "image/jpeg",
		Signatures: [][]byte{
			{0xFF, 0xD8, 0xFF, 0xDB}, // JPEG raw
			{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG/JFIF
			{0xFF, 0xD8, 0xFF, 0xE1}, // JPEG/Exif
			{0xFF, 0xD8, 0xFF, 0xE2}, // JPEG/Canon
			{0xFF, 0xD8, 0xFF, 0xE3}, // JPEG/Samsung
			{0xFF, 0xD8, 0xFF, 0xEE}, // JPEG/Adobe
		},
		Offset: 0,
	},
	{
		ContentType: "image/jpg", // Alias for JPEG
		Signatures: [][]byte{
			{0xFF, 0xD8, 0xFF, 0xDB},
			{0xFF, 0xD8, 0xFF, 0xE0},
			{0xFF, 0xD8, 0xFF, 0xE1},
			{0xFF, 0xD8, 0xFF, 0xE2},
			{0xFF, 0xD8, 0xFF, 0xE3},
			{0xFF, 0xD8, 0xFF, 0xEE},
		},
		Offset: 0,
	},
	{
		ContentType: "image/webp",
		Signatures: [][]byte{
			{0x52, 0x49, 0x46, 0x46}, // RIFF (WebP starts with RIFF)
		},
		Offset: 0,
	},
	{
		ContentType: "image/tiff",
		Signatures: [][]byte{
			{0x49, 0x49, 0x2A, 0x00}, // Little-endian TIFF
			{0x4D, 0x4D, 0x00, 0x2A}, // Big-endian TIFF
		},
		Offset: 0,
	},
	{
		ContentType: "image/svg+xml",
		Signatures: [][]byte{
			{0x3C, 0x3F, 0x78, 0x6D, 0x6C},                   // <?xml
			{0x3C, 0x73, 0x76, 0x67},                         // <svg
			{0x3C, 0x21, 0x44, 0x4F, 0x43, 0x54, 0x59, 0x50}, // <!DOCTYPE
		},
		Offset: 0,
	},

	// Video formats
	{
		ContentType: "video/mp4",
		Signatures: [][]byte{
			{0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6F, 0x6D}, // ftypisom (at offset 4)
			{0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32}, // ftypmp42
			{0x66, 0x74, 0x79, 0x70, 0x4D, 0x53, 0x4E, 0x56}, // ftypMSNV
			{0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x31}, // ftypmp41
		},
		Offset: 4,
	},
	{
		ContentType: "video/mpeg",
		Signatures: [][]byte{
			{0x00, 0x00, 0x01, 0xBA}, // MPEG Program Stream
			{0x00, 0x00, 0x01, 0xB3}, // MPEG video stream
		},
		Offset: 0,
	},
	{
		ContentType: "video/ogg",
		Signatures: [][]byte{
			{0x4F, 0x67, 0x67, 0x53}, // OggS
		},
		Offset: 0,
	},

	// Document formats
	{
		ContentType: "application/pdf",
		Signatures: [][]byte{
			{0x25, 0x50, 0x44, 0x46}, // %PDF
		},
		Offset: 0,
	},
	// text/plain has no magic bytes - we'll allow it through
}

// ValidateFileSignature checks if file's magic bytes match its claimed content type
func ValidateFileSignature(fileHeader *multipart.FileHeader, contentType string) error {
	// For text/plain, we skip magic byte validation as plain text has no signature
	if contentType == "text/plain" {
		return nil
	}

	// Get file bytes (read first 512 bytes which is enough for all signatures)
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 512 bytes for signature detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return fmt.Errorf("failed to read file: %w", err)
	}
	buffer = buffer[:n]

	// Find matching signature for content type
	var matchingSignature *FileSignature
	for i := range fileSignatures {
		if fileSignatures[i].ContentType == contentType {
			matchingSignature = &fileSignatures[i]
			break
		}
	}

	if matchingSignature == nil {
		return fmt.Errorf("%w: %s", ErrUnsupportedFileType, contentType)
	}

	// Check if file has enough bytes for signature at offset
	if len(buffer) < matchingSignature.Offset+4 {
		return fmt.Errorf("%w: file too small", ErrInvalidFileSignature)
	}

	// Check if any signature matches
	for _, signature := range matchingSignature.Signatures {
		sigLen := len(signature)
		if matchingSignature.Offset+sigLen > len(buffer) {
			continue
		}

		fileBytes := buffer[matchingSignature.Offset : matchingSignature.Offset+sigLen]
		if bytes.Equal(fileBytes, signature) {
			// Additional check for WebP - verify WEBP string at offset 8
			if contentType == "image/webp" {
				if len(buffer) >= 12 {
					webpMarker := buffer[8:12]
					if bytes.Equal(webpMarker, []byte{0x57, 0x45, 0x42, 0x50}) {
						return nil
					}
				}
				continue // Try next signature if WEBP marker not found
			}
			return nil // Valid signature found
		}
	}

	return fmt.Errorf("%w: expected %s but file signature doesn't match",
		ErrInvalidFileSignature, contentType)
}

// ValidateFileExtension checks if file extension matches content type
func ValidateFileExtension(filename string, contentType string) error {
	expectedExtensions := map[string][]string{
		"image/png":       {".png"},
		"image/jpeg":      {".jpg", ".jpeg"},
		"image/jpg":       {".jpg", ".jpeg"},
		"image/svg+xml":   {".svg"},
		"image/webp":      {".webp"},
		"image/tiff":      {".tiff", ".tif"},
		"video/mp4":       {".mp4"},
		"video/mpeg":      {".mpeg", ".mpg"},
		"video/ogg":       {".ogg", ".ogv"},
		"application/pdf": {".pdf"},
		"text/plain":      {".txt"},
	}

	extensions, ok := expectedExtensions[contentType]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnsupportedFileType, contentType)
	}

	// Get file extension (case-insensitive)
	for _, ext := range extensions {
		if len(filename) >= len(ext) {
			fileExt := filename[len(filename)-len(ext):]
			if bytes.EqualFold([]byte(fileExt), []byte(ext)) {
				return nil
			}
		}
	}

	return fmt.Errorf("file extension doesn't match content type %s (expected one of %v)",
		contentType, extensions)
}

// ValidateFile performs comprehensive file validation
func ValidateFile(fileHeader *multipart.FileHeader, contentType string) error {
	// 1. Validate file extension
	if err := ValidateFileExtension(fileHeader.Filename, contentType); err != nil {
		return err
	}

	// 2. Validate magic bytes
	if err := ValidateFileSignature(fileHeader, contentType); err != nil {
		return err
	}

	return nil
}
