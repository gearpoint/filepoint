package views

// WebhookPayload contains the webhook request body.
type WebhookPayload struct {
	Id            string `json:"id"`
	Success       bool   `json:"success"`
	CorrelationId string `json:"correlationId"`
	Location      string `json:"location"`
	Error         string `json:"error"`
}
