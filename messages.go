package arara

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

const (
	messagesBase         = "/v1/messages"
	idempotencyKeyHeader = "Idempotency-Key"
	// MaxBatchSize is the maximum number of items accepted by Messages.SendBatch.
	MaxBatchSize = 1000
)

// MessagesService handles the /v1/messages resource.
type MessagesService struct {
	client *Client
}

// SendMessageRequest is the payload for sending a WhatsApp message.
// Receiver accepts "whatsapp:+5511...", "+5511..." or digits only.
type SendMessageRequest struct {
	Receiver          string         `json:"receiver"`
	Sender            string         `json:"sender,omitempty"`
	Type              string         `json:"type,omitempty"`
	TemplateName      string         `json:"templateName,omitempty"`
	TemplateVariables []string       `json:"templateVariables,omitempty"`
	SmartLinkParam    string         `json:"smartLinkParam,omitempty"`
	SmartLinkURL      string         `json:"smartLinkUrl,omitempty"`
	Body              string         `json:"body,omitempty"`
	Interactive       map[string]any `json:"interactive,omitempty"`
	Charge            *ChargePayload `json:"charge,omitempty"`
	Location          map[string]any `json:"location,omitempty"`
	Reaction          map[string]any `json:"reaction,omitempty"`
	ReplyTo           string         `json:"replyTo,omitempty"`
	ScheduledAt       string         `json:"scheduled_at,omitempty"`
	Mode              string         `json:"mode,omitempty"`
	// Deprecated: the API removes media_url on 2027-01-01. Use a template with a media header.
	MediaURL string `json:"media_url,omitempty"`
}

// ChargePayload is a payment card sent in the conversation. Amounts are in cents and exactly
// one payment method (Pix, PaymentLink or Boleto) is expected.
type ChargePayload struct {
	ReferenceID   string         `json:"referenceId"`
	Items         []ChargeItem   `json:"items"`
	Pix           *PixPayment    `json:"pix,omitempty"`
	PaymentLink   *PaymentLink   `json:"paymentLink,omitempty"`
	Boleto        *BoletoPayment `json:"boleto,omitempty"`
	ShippingCents int64          `json:"shippingCents,omitempty"`
	DiscountCents int64          `json:"discountCents,omitempty"`
	// ExpiresAt is epoch seconds; Meta requires at least 5 minutes ahead.
	ExpiresAt *int64 `json:"expiresAt,omitempty"`
	ImageURL  string `json:"imageUrl,omitempty"`
}

// ChargeItem is a line of a charge.
type ChargeItem struct {
	Name        string `json:"name"`
	AmountCents int64  `json:"amountCents"`
	Quantity    int    `json:"quantity,omitempty"`
	ID          string `json:"id,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
}

// PixPayment is a Pix payment method. KeyType is one of CPF, CNPJ, EMAIL, PHONE or EVP.
type PixPayment struct {
	Code         string `json:"code"`
	Key          string `json:"key"`
	KeyType      string `json:"keyType"`
	MerchantName string `json:"merchantName"`
}

// PaymentLink is a hosted payment link.
type PaymentLink struct {
	URL string `json:"url"`
}

// BoletoPayment is a boleto payment method.
type BoletoPayment struct {
	DigitableLine string `json:"digitableLine"`
}

// SendOptions carries per-call options for sending messages.
type SendOptions struct {
	// IdempotencyKey is sent as the Idempotency-Key header. A UUID v4 is generated when empty
	// and reused on every retry of the same call.
	IdempotencyKey string
}

// MessageResponse is a message as returned by the API.
type MessageResponse struct {
	ID       *string  `json:"id"`
	Status   string   `json:"status"`
	Mode     string   `json:"mode"`
	Sender   string   `json:"sender"`
	Receiver string   `json:"receiver"`
	Body     *string  `json:"body"`
	Cost     *float64 `json:"cost"`
	Reason   *string  `json:"reason"`
}

// BatchMessageItem is a single recipient of a batch send.
type BatchMessageItem struct {
	Receiver          string   `json:"receiver"`
	TemplateVariables []string `json:"templateVariables,omitempty"`
	SmartLinkParam    string   `json:"smartLinkParam,omitempty"`
	SmartLinkURL      string   `json:"smartLinkUrl,omitempty"`
	MediaURL          string   `json:"mediaUrl,omitempty"`
}

// BatchMessageRequest is the payload for a batch send (up to MaxBatchSize items).
type BatchMessageRequest struct {
	TemplateName string             `json:"templateName"`
	Messages     []BatchMessageItem `json:"messages"`
}

// BatchMessageItemResponse is the per-recipient result of a batch send.
type BatchMessageItemResponse struct {
	ID       *string  `json:"id"`
	Receiver string   `json:"receiver"`
	Status   string   `json:"status"`
	Cost     *float64 `json:"cost"`
}

// BatchMessageResponse is the result of a batch send.
type BatchMessageResponse struct {
	BatchID      string                     `json:"batchId"`
	TemplateName string                     `json:"templateName"`
	Total        int                        `json:"total"`
	Accepted     int                        `json:"accepted"`
	TotalCost    float64                    `json:"totalCost"`
	Messages     []BatchMessageItemResponse `json:"messages"`
}

// Send sends a WhatsApp message. POST /v1/messages
func (s *MessagesService) Send(ctx context.Context, req *SendMessageRequest, opts ...SendOptions) (*MessageResponse, error) {
	var out MessageResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: messagesBase, body: req, headers: idempotencyHeaders(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SendBatch sends one template to up to MaxBatchSize recipients. POST /v1/messages/batch
func (s *MessagesService) SendBatch(ctx context.Context, req *BatchMessageRequest, opts ...SendOptions) (*BatchMessageResponse, error) {
	if err := validateBatch(req); err != nil {
		return nil, err
	}
	var out BatchMessageResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: messagesBase + "/batch", body: req, headers: idempotencyHeaders(opts)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a message by id. GET /v1/messages/{id}
func (s *MessagesService) Get(ctx context.Context, id string) (*MessageResponse, error) {
	var out MessageResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: messagesBase + "/" + url.PathEscape(id)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListByBatch lists the messages of a batch. GET /v1/messages?batchId=
func (s *MessagesService) ListByBatch(ctx context.Context, batchID string) ([]MessageResponse, error) {
	var out []MessageResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: messagesBase, query: url.Values{"batchId": {batchID}}}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func idempotencyHeaders(opts []SendOptions) map[string]string {
	key := ""
	if len(opts) > 0 {
		key = strings.TrimSpace(opts[0].IdempotencyKey)
	}
	if key == "" {
		key = newUUIDv4()
	}
	return map[string]string{idempotencyKeyHeader: key}
}

func validateBatch(req *BatchMessageRequest) *APIError {
	switch {
	case req == nil:
		return invalidBatch("batch request is required")
	case strings.TrimSpace(req.TemplateName) == "":
		return invalidBatch("templateName is required")
	case len(req.Messages) == 0:
		return invalidBatch("a batch needs at least one message")
	case len(req.Messages) > MaxBatchSize:
		return &APIError{Code: "BATCH_TOO_LARGE", Message: "a batch accepts at most 1000 messages"}
	default:
		return nil
	}
}

func invalidBatch(message string) *APIError {
	return &APIError{Code: codeInvalidRequest, Message: message}
}
