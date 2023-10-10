package views

// WebhookPayload contains the webhook request body.
type WebhookPayload struct {
	Id       string   `json:"id"`
	Success  bool     `json:"success"`
	Location string   `json:"location"`
	Labels   []string `json:"labels"`
	Error    string   `json:"error"`
}
