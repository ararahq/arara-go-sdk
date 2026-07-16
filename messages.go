package arara

import (
	"context"
	"net/http"
)

// MessagesService handles the /v1/messages resource.
type MessagesService struct {
	client *Client
}

// SendMessageRequest is the payload for sending a WhatsApp message.
type SendMessageRequest struct {
	Receiver          string   `json:"receiver"`
	Sender            string   `json:"sender,omitempty"`
	Type              string   `json:"type,omitempty"`
	TemplateName      string   `json:"templateName,omitempty"`
	TemplateVariables []string `json:"templateVariables,omitempty"`
	SmartLinkParam    string   `json:"smartLinkParam,omitempty"`
	SmartLinkURL      string   `json:"smartLinkUrl,omitempty"`
	Body              string   `json:"body,omitempty"`
	ReplyTo           string   `json:"replyTo,omitempty"`
	ScheduledAt       string   `json:"scheduled_at,omitempty"`
	Mode              string   `json:"mode,omitempty"`
	MediaURL          string   `json:"media_url,omitempty"`
}

// SendOptions carries per-call options for sending a message.
type SendOptions struct {
	// IdempotencyKey is sent as the Idempotency-Key header to deduplicate retried sends.
	IdempotencyKey string
}

// MessageResponse is the response for a sent message.
type MessageResponse struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Mode     string   `json:"mode"`
	Sender   string   `json:"sender"`
	Receiver string   `json:"receiver"`
	Body     *string  `json:"body"`
	Cost     *float64 `json:"cost"`
}

// Send sends a WhatsApp message. POST /v1/messages
func (s *MessagesService) Send(ctx context.Context, req *SendMessageRequest, opts ...SendOptions) (*MessageResponse, error) {
	headers := map[string]string{}
	if len(opts) > 0 && opts[0].IdempotencyKey != "" {
		headers["Idempotency-Key"] = opts[0].IdempotencyKey
	}
	var out MessageResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: "/v1/messages", body: req, headers: headers}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
