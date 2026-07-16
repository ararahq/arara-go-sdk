package arara

import (
	"context"
	"net/http"
)

// SmartLinksService handles the /v1/smart-links/whatsapp resource.
type SmartLinksService struct {
	client *Client
}

// CreateWhatsAppSmartLinkRequest is the payload for creating a smart link.
type CreateWhatsAppSmartLinkRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	DefaultText string `json:"defaultText,omitempty"`
	QRCodeColor string `json:"qrCodeColor,omitempty"`
}

// UpdateWhatsAppSmartLinkRequest is the payload for updating a smart link.
type UpdateWhatsAppSmartLinkRequest struct {
	Name        string `json:"name,omitempty"`
	DefaultText string `json:"defaultText,omitempty"`
	QRCodeColor string `json:"qrCodeColor,omitempty"`
}

// WhatsAppSmartLinkResponse is a smart link record.
type WhatsAppSmartLinkResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phoneNumber"`
	DefaultText *string `json:"defaultText"`
	QRCodeColor string  `json:"qrCodeColor"`
	Code        string  `json:"code"`
	ShortURL    string  `json:"shortUrl"`
	CreatedAt   *string `json:"createdAt"`
	Clicks      int64   `json:"clicks"`
}

const smartLinksBase = "/v1/smart-links/whatsapp"

// Create creates a WhatsApp smart link. POST /v1/smart-links/whatsapp
func (s *SmartLinksService) Create(ctx context.Context, req *CreateWhatsAppSmartLinkRequest) (*WhatsAppSmartLinkResponse, error) {
	var out WhatsAppSmartLinkResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: smartLinksBase, body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a WhatsApp smart link. PUT /v1/smart-links/whatsapp/{id}
func (s *SmartLinksService) Update(ctx context.Context, id string, req *UpdateWhatsAppSmartLinkRequest) (*WhatsAppSmartLinkResponse, error) {
	var out WhatsAppSmartLinkResponse
	err := s.client.do(ctx, request{method: http.MethodPut, path: smartLinksBase + "/" + id, body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List lists WhatsApp smart links. GET /v1/smart-links/whatsapp
func (s *SmartLinksService) List(ctx context.Context) ([]WhatsAppSmartLinkResponse, error) {
	var out []WhatsAppSmartLinkResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: smartLinksBase}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Stats returns smart link click stats. GET /v1/smart-links/whatsapp/{id}/stats
func (s *SmartLinksService) Stats(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: smartLinksBase + "/" + id + "/stats"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
