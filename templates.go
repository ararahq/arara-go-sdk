package arara

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
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
	Name             string           `json:"name"`
	Category         string           `json:"category"`
	Language         string           `json:"language"`
	Body             string           `json:"body"`
	Header           string           `json:"header,omitempty"`
	HeaderType       string           `json:"headerType,omitempty"`
	Footer           string           `json:"footer,omitempty"`
	Buttons          []TemplateButton `json:"buttons,omitempty"`
	Samples          map[string]any   `json:"samples,omitempty"`
	VariableExamples []string         `json:"variableExamples,omitempty"`
}

// Template is a WhatsApp message template, mirroring the API's TemplateResponse.
type Template struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	FormattedName       string            `json:"formattedName"`
	Category            string            `json:"category"`
	OriginalCategory    *string           `json:"originalCategory"`
	Language            string            `json:"language"`
	ProviderName        string            `json:"providerName"`
	ProviderTemplateID  string            `json:"providerTemplateId"`
	ProviderStatus      string            `json:"providerStatus"`
	RejectionReason     *string           `json:"rejectionReason"`
	AvailableForSending bool              `json:"availableForSending"`
	UnavailableReason   *string           `json:"unavailableReason"`
	BodyPreview         *string           `json:"bodyPreview"`
	StructureJSON       json.RawMessage   `json:"structureJson"`
	UsageGuide          map[string]any    `json:"usageGuide"`
	VariablesSchema     map[string]string `json:"variablesSchema"`
	CreatedAt           string            `json:"createdAt"`
	UpdatedAt           *string           `json:"updatedAt"`
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
	Category        *string `json:"category"`
}

// TemplateListParams are the optional filters for listing templates.
type TemplateListParams struct {
	Name   string
	Status string
	Page   int
	Size   int
}

const (
	templatesBase            = "/v1/templates"
	defaultTemplatesPageSize = 50
	defaultAnalyticsPeriod   = "30d"
)

// List lists templates, paginated. GET /v1/templates
func (s *TemplatesService) List(ctx context.Context, params TemplateListParams) (*Paginated[Template], error) {
	q := PageParams{Page: params.Page, Size: params.Size}.values(defaultTemplatesPageSize)
	if params.Name != "" {
		q.Set("name", params.Name)
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	var out Paginated[Template]
	err := s.client.do(ctx, request{method: http.MethodGet, path: templatesBase, query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// FindByName returns the first template on the first page whose name matches exactly, using
// List with the name filter. It returns a 404 NOT_FOUND *APIError when there is no match.
func (s *TemplatesService) FindByName(ctx context.Context, name string) (*Template, error) {
	page, err := s.List(ctx, TemplateListParams{Name: name})
	if err != nil {
		return nil, err
	}
	for i := range page.Data {
		if page.Data[i].Name == name {
			return &page.Data[i], nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND", Message: "template not found: " + name}
}

// Create creates a new template for Meta approval. POST /v1/templates
func (s *TemplatesService) Create(ctx context.Context, req *CreateTemplateRequest) (*TemplateResponse, error) {
	var out TemplateResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: templatesBase, body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a template by id (UUID). GET /v1/templates/{id}
func (s *TemplatesService) Get(ctx context.Context, id string) (*Template, error) {
	var out Template
	err := s.client.do(ctx, request{method: http.MethodGet, path: templatePath(id)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStatus retrieves a template's provider status by id (UUID). GET /v1/templates/{id}/status
func (s *TemplatesService) GetStatus(ctx context.Context, id string) (*TemplateStatus, error) {
	var out TemplateStatus
	err := s.client.do(ctx, request{method: http.MethodGet, path: templatePath(id) + "/status"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a template by id (UUID). DELETE /v1/templates/{id}
func (s *TemplatesService) Delete(ctx context.Context, id string) error {
	return s.client.do(ctx, request{method: http.MethodDelete, path: templatePath(id)}, nil)
}

// Analytics returns aggregate template analytics. GET /v1/templates/analytics
// Period defaults to "30d".
func (s *TemplatesService) Analytics(ctx context.Context, period string) (map[string]any, error) {
	return s.analytics(ctx, templatesBase+"/analytics", period)
}

// TemplateAnalytics returns analytics for one template by id (UUID). GET /v1/templates/{id}/analytics
func (s *TemplatesService) TemplateAnalytics(ctx context.Context, id, period string) (map[string]any, error) {
	return s.analytics(ctx, templatePath(id)+"/analytics", period)
}

func (s *TemplatesService) analytics(ctx context.Context, path, period string) (map[string]any, error) {
	if period == "" {
		period = defaultAnalyticsPeriod
	}
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: path, query: url.Values{"period": {period}}}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func templatePath(id string) string {
	return templatesBase + "/" + url.PathEscape(id)
}
