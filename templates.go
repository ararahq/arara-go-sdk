package arara

import (
	"context"
	"net/http"
)

// TemplatesService handles the /v1/templates resource.
type TemplatesService struct {
	client *Client
}

// TemplateButton describes a template button.
type TemplateButton struct {
	Type        string         `json:"type"`
	Text        string         `json:"text"`
	URL         string         `json:"url,omitempty"`
	Phone       string         `json:"phone,omitempty"`
	ExtraConfig map[string]any `json:"extraConfig,omitempty"`
}

// CreateTemplateRequest is the payload for creating a template.
type CreateTemplateRequest struct {
	Name             string            `json:"name"`
	Category         string            `json:"category"`
	Language         string            `json:"language"`
	Body             string            `json:"body"`
	Header           string            `json:"header,omitempty"`
	HeaderType       string            `json:"headerType,omitempty"`
	Footer           string            `json:"footer,omitempty"`
	Buttons          []TemplateButton  `json:"buttons,omitempty"`
	Samples          map[string]string `json:"samples,omitempty"`
	VariableExamples []string          `json:"variableExamples,omitempty"`
}

// Template is a WhatsApp message template.
type Template struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	FormattedName   string   `json:"formattedName"`
	Category        string   `json:"category"`
	Language        string   `json:"language"`
	Body            string   `json:"body"`
	Samples         []string `json:"samples"`
	ButtonsConfig   []any    `json:"buttonsConfig"`
	ProviderStatus  string   `json:"providerStatus"`
	RejectionReason *string  `json:"rejectionReason"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       *string  `json:"updatedAt"`
}

// TemplateResponse is the response for a created template.
type TemplateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TemplateStatus reports the provider approval status of a template.
type TemplateStatus struct {
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejectionReason"`
	Category        string  `json:"category"`
}

// List lists all templates. GET /v1/templates
func (s *TemplatesService) List(ctx context.Context) ([]Template, error) {
	var out []Template
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/templates"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a new template for Meta approval. POST /v1/templates
func (s *TemplatesService) Create(ctx context.Context, req *CreateTemplateRequest) (*TemplateResponse, error) {
	var out TemplateResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: "/v1/templates", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a template by name. GET /v1/templates/{name}
func (s *TemplatesService) Get(ctx context.Context, name string) (*Template, error) {
	var out Template
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/templates/" + name}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStatus retrieves a template's provider status. GET /v1/templates/{name}/status
func (s *TemplatesService) GetStatus(ctx context.Context, name string) (*TemplateStatus, error) {
	var out TemplateStatus
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/templates/" + name + "/status"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a template by name. DELETE /v1/templates/{name}
func (s *TemplatesService) Delete(ctx context.Context, name string) error {
	return s.client.do(ctx, request{method: http.MethodDelete, path: "/v1/templates/" + name}, nil)
}
