package views

import (
	"time"
)

// UploadPubSub contains the view used in pub/sub.
type UploadPubSub struct {
	Id          string    `validate:"required,uuid" json:"id"`
	UserId      string    `validate:"required,uuid" json:"userId"`
	Author      string    `validate:"omitempty,min=4,max=30" json:"author"`
	Title       string    `validate:"omitempty,min=4,max=100" json:"title"`
	Filename    string    `validate:"required" json:"filename"`
	ContentType string    `validate:"required" json:"contentType"`
	Size        int64     `validate:"required,max-file-size" json:"size"`
	IpAddress   string    `validate:"required" json:"ip"`
	OccurredOn  time.Time `validate:"required" json:"occurredOn"`
}
