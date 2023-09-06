package views

import "time"

// Upload contains the upload view.
type Upload struct {
	Id         string    `json:"id,omitempty"`
	Author     string    `json:"author"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	OccurredOn time.Time `json:"occurred_on"`
}
