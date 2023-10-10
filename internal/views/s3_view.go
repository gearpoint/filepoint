package views

import "io"

// S3UploadInput is the s3 upload view.
type S3UploadInput struct {
	File        io.Reader
	Name        string
	Size        int64
	ContentType string
	BucketName  string
}
