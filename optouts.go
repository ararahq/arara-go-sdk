package arara

import (
	"context"
	"net/http"
	"net/url"
)

// OptOutsService handles the /v1/opt-outs resource. Requires an ADMIN API key.
type OptOutsService struct {
	client *Client
}

// OptOutRequest registers a phone as opted out.
type OptOutRequest struct {
	Phone  string `json:"phone"`
	Reason string `json:"reason,omitempty"`
}

const optOutsBase = "/v1/opt-outs"

// List lists opted-out phones. GET /v1/opt-outs
func (s *OptOutsService) List(ctx context.Context) (map[string]any, error) {
	return s.call(ctx, http.MethodGet, optOutsBase, nil)
}

// Get returns the opt-out state of a phone. GET /v1/opt-outs/{phone}
func (s *OptOutsService) Get(ctx context.Context, phone string) (map[string]any, error) {
	return s.call(ctx, http.MethodGet, optOutsBase+"/"+url.PathEscape(phone), nil)
}

// Create registers an opt-out. POST /v1/opt-outs
func (s *OptOutsService) Create(ctx context.Context, req *OptOutRequest) (map[string]any, error) {
	return s.call(ctx, http.MethodPost, optOutsBase, req)
}

// Delete removes an opt-out. DELETE /v1/opt-outs/{phone}
func (s *OptOutsService) Delete(ctx context.Context, phone string) (map[string]any, error) {
	return s.call(ctx, http.MethodDelete, optOutsBase+"/"+url.PathEscape(phone), nil)
}

func (s *OptOutsService) call(ctx context.Context, method, path string, body any) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: method, path: path, body: body}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
