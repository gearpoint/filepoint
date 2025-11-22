package utils

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createMockFileHeader creates a mock multipart.FileHeader for testing
func createMockFileHeader(filename string, content []byte) *multipart.FileHeader {
	header := &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len(content)),
		Header:   make(textproto.MIMEHeader),
	}

	// Create a pipe to simulate file content
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		part, _ := writer.CreateFormFile("file", filename)
		part.Write(content)
		writer.Close()
	}()

	reader := multipart.NewReader(pr, writer.Boundary())
	part, _ := reader.NextPart()

	// Read the content back
	buf := new(bytes.Buffer)
	buf.ReadFrom(part)

	// Store content for Open() method
	header.Header = part.Header

	return header
}

// createTestFileHeader creates a proper test file header with content
func createTestFileHeader(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(int64(len(content)) + 1024)

	if files, ok := form.File["file"]; ok && len(files) > 0 {
		return files[0]
	}

	return nil
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		expectError bool
	}{
		// Valid cases
		{"Valid PNG", "image.png", "image/png", false},
		{"Valid JPEG", "photo.jpg", "image/jpeg", false},
		{"Valid JPEG alt", "photo.jpeg", "image/jpeg", false},
		{"Valid WebP", "image.webp", "image/webp", false},
		{"Valid SVG", "icon.svg", "image/svg+xml", false},
		{"Valid TIFF", "scan.tiff", "image/tiff", false},
		{"Valid TIFF alt", "scan.tif", "image/tiff", false},
		{"Valid MP4", "video.mp4", "video/mp4", false},
		{"Valid MPEG", "movie.mpeg", "video/mpeg", false},
		{"Valid OGG", "video.ogg", "video/ogg", false},
		{"Valid PDF", "document.pdf", "application/pdf", false},
		{"Valid TXT", "readme.txt", "text/plain", false},

		// Case insensitive
		{"Valid PNG uppercase", "IMAGE.PNG", "image/png", false},
		{"Valid JPG mixed case", "Photo.JpG", "image/jpeg", false},

		// Invalid cases
		{"Wrong extension for PNG", "image.jpg", "image/png", true},
		{"Wrong extension for PDF", "doc.txt", "application/pdf", true},
		{"Wrong extension for MP4", "video.avi", "video/mp4", true},
		{"Unsupported type", "file.exe", "application/octet-stream", true},
		{"No extension", "file", "image/png", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileExtension(tt.filename, tt.contentType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFileSignature(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		content     []byte
		expectError bool
	}{
		// Valid signatures
		{
			name:        "Valid PNG",
			filename:    "test.png",
			contentType: "image/png",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			expectError: false,
		},
		{
			name:        "Valid JPEG JFIF",
			filename:    "test.jpg",
			contentType: "image/jpeg",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			expectError: false,
		},
		{
			name:        "Valid JPEG Exif",
			filename:    "photo.jpg",
			contentType: "image/jpeg",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE1, 0x00, 0x00, 0x45, 0x78, 0x69, 0x66},
			expectError: false,
		},
		{
			name:        "Valid WebP",
			filename:    "image.webp",
			contentType: "image/webp",
			content:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			expectError: false,
		},
		{
			name:        "Valid TIFF little-endian",
			filename:    "scan.tiff",
			contentType: "image/tiff",
			content:     []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00},
			expectError: false,
		},
		{
			name:        "Valid TIFF big-endian",
			filename:    "scan.tif",
			contentType: "image/tiff",
			content:     []byte{0x4D, 0x4D, 0x00, 0x2A, 0x00, 0x00, 0x00, 0x08},
			expectError: false,
		},
		{
			name:        "Valid SVG with XML declaration",
			filename:    "icon.svg",
			contentType: "image/svg+xml",
			content:     []byte("<?xml version=\"1.0\"?><svg></svg>"),
			expectError: false,
		},
		{
			name:        "Valid SVG with svg tag",
			filename:    "graphic.svg",
			contentType: "image/svg+xml",
			content:     []byte("<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>"),
			expectError: false,
		},
		{
			name:        "Valid PDF",
			filename:    "doc.pdf",
			contentType: "application/pdf",
			content:     []byte("%PDF-1.4\n%����\n"),
			expectError: false,
		},
		{
			name:        "Valid plain text",
			filename:    "readme.txt",
			contentType: "text/plain",
			content:     []byte("This is a plain text file"),
			expectError: false,
		},
		{
			name:        "Valid MPEG Program Stream",
			filename:    "video.mpeg",
			contentType: "video/mpeg",
			content:     []byte{0x00, 0x00, 0x01, 0xBA, 0x44, 0x00, 0x04, 0x00},
			expectError: false,
		},
		{
			name:        "Valid OGG",
			filename:    "video.ogg",
			contentType: "video/ogg",
			content:     []byte{0x4F, 0x67, 0x67, 0x53, 0x00, 0x02, 0x00, 0x00},
			expectError: false,
		},

		// Invalid signatures
		{
			name:        "Invalid PNG signature",
			filename:    "fake.png",
			contentType: "image/png",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}, // JPEG bytes
			expectError: true,
		},
		{
			name:        "Invalid JPEG signature",
			filename:    "fake.jpg",
			contentType: "image/jpeg",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG bytes
			expectError: true,
		},
		{
			name:        "Invalid WebP - missing WEBP marker",
			filename:    "fake.webp",
			contentType: "image/webp",
			content:     []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x46, 0x41, 0x4B, 0x45},
			expectError: true,
		},
		{
			name:        "Invalid PDF signature",
			filename:    "fake.pdf",
			contentType: "application/pdf",
			content:     []byte("This is not a PDF file"),
			expectError: true,
		},
		{
			name:        "File too small",
			filename:    "tiny.png",
			contentType: "image/png",
			content:     []byte{0x89},
			expectError: true,
		},
		{
			name:        "Empty file",
			filename:    "empty.jpg",
			contentType: "image/jpeg",
			content:     []byte{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := createTestFileHeader(tt.filename, tt.content)
			if fileHeader == nil {
				t.Fatal("Failed to create test file header")
			}

			err := ValidateFileSignature(fileHeader, tt.contentType)
			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.name)
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.name)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		content     []byte
		expectError bool
	}{
		{
			name:        "Valid PNG file",
			filename:    "image.png",
			contentType: "image/png",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			expectError: false,
		},
		{
			name:        "Valid JPEG file",
			filename:    "photo.jpg",
			contentType: "image/jpeg",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46},
			expectError: false,
		},
		{
			name:        "Wrong extension",
			filename:    "image.jpg",
			contentType: "image/png",
			content:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expectError: true,
		},
		{
			name:        "Wrong signature",
			filename:    "image.png",
			contentType: "image/png",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46},
			expectError: true,
		},
		{
			name:        "Valid PDF",
			filename:    "document.pdf",
			contentType: "application/pdf",
			content:     []byte("%PDF-1.5\n"),
			expectError: false,
		},
		{
			name:        "Unsupported type",
			filename:    "app.exe",
			contentType: "application/octet-stream",
			content:     []byte{0x4D, 0x5A}, // EXE signature
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := createTestFileHeader(tt.filename, tt.content)
			if fileHeader == nil {
				t.Fatal("Failed to create test file header")
			}

			err := ValidateFile(fileHeader, tt.contentType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMP4Signature(t *testing.T) {
	// MP4 has signature at offset 4
	validMP4 := []byte{
		0x00, 0x00, 0x00, 0x20, // Box size
		0x66, 0x74, 0x79, 0x70, // 'ftyp'
		0x69, 0x73, 0x6F, 0x6D, // 'isom'
	}

	fileHeader := createTestFileHeader("video.mp4", validMP4)
	if fileHeader == nil {
		t.Fatal("Failed to create test file header")
	}

	err := ValidateFileSignature(fileHeader, "video/mp4")
	assert.NoError(t, err, "Valid MP4 should pass validation")

	// Test invalid MP4
	invalidMP4 := []byte{0x00, 0x00, 0x00, 0x20, 0x46, 0x41, 0x4B, 0x45}
	fileHeader = createTestFileHeader("fake.mp4", invalidMP4)
	if fileHeader == nil {
		t.Fatal("Failed to create test file header")
	}

	err = ValidateFileSignature(fileHeader, "video/mp4")
	assert.Error(t, err, "Invalid MP4 should fail validation")
}
