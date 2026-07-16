package arara

import (
	"context"
	"net/http"
)

// OrganizationsService handles organization webhook configuration.
type OrganizationsService struct {
	client *Client
}

// UpdateWebhookRequest is the payload for updating the organization webhook.
type UpdateWebhookRequest struct {
	URL            string `json:"url,omitempty"`
	Secret         string `json:"secret,omitempty"`
	GenerateSecret *bool  `json:"generateSecret,omitempty"`
}

// GetWebhook returns the organization webhook configuration. GET /organizations/me/webhook
func (s *OrganizationsService) GetWebhook(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/organizations/me/webhook"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateWebhook updates the organization webhook configuration. PATCH /organizations/me/webhook
func (s *OrganizationsService) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodPatch, path: "/organizations/me/webhook", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
