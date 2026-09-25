package arara

import (
	"context"
	"net/http"
	"net/url"
)

// NumbersService handles the /v1/organizations/me/numbers resource.
type NumbersService struct {
	client *Client
}

// NumberCardDTO describes a single WhatsApp number and its health.
type NumberCardDTO struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Alias             *string `json:"alias"`
	Description       *string `json:"description"`
	PhoneNumber       string  `json:"phoneNumber"`
	Type              string  `json:"type"`
	IsDefault         bool    `json:"isDefault"`
	Status            string  `json:"status"`
	QualityScore      string  `json:"qualityScore"`
	MessagingTier     string  `json:"messagingTier"`
	VerifiedAt        *string `json:"verifiedAt"`
	LastHealthCheckAt *string `json:"lastHealthCheckAt"`
	Provider          string  `json:"provider"`
	CreatedAt         *string `json:"createdAt"`
	MessagesLast7d    int64   `json:"messagesLast7d"`
	MessagesLast30d   int64   `json:"messagesLast30d"`
}

// NumbersSlotDTO describes plan slot usage for numbers.
type NumbersSlotDTO struct {
	Used              int    `json:"used"`
	Max               int    `json:"max"`
	PlanLabel         string `json:"planLabel"`
	AtCap             bool   `json:"atCap"`
	NoEntitlement     bool   `json:"noEntitlement"`
	MonthlyPriceCents int    `json:"monthlyPriceCents"`
	MonthlyTotalCents int    `json:"monthlyTotalCents"`
}

// NumbersResponseDTO is the numbers listing with plan slot info.
type NumbersResponseDTO struct {
	Numbers []NumberCardDTO `json:"numbers"`
	Slot    NumbersSlotDTO  `json:"slot"`
}

// UpdateNumberRequest is the payload for updating a number.
type UpdateNumberRequest struct {
	Alias       string `json:"alias,omitempty"`
	IsDefault   *bool  `json:"isDefault,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// RequestNumberRequest is the payload for requesting a new dedicated number.
type RequestNumberRequest struct {
	Reason            string `json:"reason,omitempty"`
	ExpectedVolume    string `json:"expectedVolume,omitempty"`
	AreaCode          string `json:"areaCode,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	ProfilePictureURL string `json:"profilePictureUrl,omitempty"`
}

const numbersBase = "/v1/organizations/me/numbers"

// List lists numbers with plan slot info. GET /v1/organizations/me/numbers
func (s *NumbersService) List(ctx context.Context) (*NumbersResponseDTO, error) {
	var out NumbersResponseDTO
	err := s.client.do(ctx, request{method: http.MethodGet, path: numbersBase}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a number. PATCH /v1/organizations/me/numbers/{id}
func (s *NumbersService) Update(ctx context.Context, id string, req *UpdateNumberRequest) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodPatch, path: numbersBase + "/" + url.PathEscape(id), body: req}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete deactivates a number. DELETE /v1/organizations/me/numbers/{id}
func (s *NumbersService) Delete(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodDelete, path: numbersBase + "/" + url.PathEscape(id)}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Request requests a new dedicated number. POST /v1/organizations/me/numbers/request
func (s *NumbersService) Request(ctx context.Context, req *RequestNumberRequest) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodPost, path: numbersBase + "/request", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListRequests lists number provisioning requests. GET /v1/organizations/me/numbers/requests
func (s *NumbersService) ListRequests(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: numbersBase + "/requests"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Sync syncs a number's health from the provider. POST /v1/organizations/me/numbers/{id}/sync
func (s *NumbersService) Sync(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodPost, path: numbersBase + "/" + url.PathEscape(id) + "/sync"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Warming returns warming recommendations for a number. GET /v1/organizations/me/numbers/{id}/warming
func (s *NumbersService) Warming(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: numbersBase + "/" + url.PathEscape(id) + "/warming"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
