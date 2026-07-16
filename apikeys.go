package arara

import (
	"context"
	"net/http"
	"net/url"
)

// APIKeysService handles the /v1/api-keys resource.
type APIKeysService struct {
	client *Client
}

// APIKey is a masked API key record.
type APIKey struct {
	ID         string `json:"id"`
	Prefix     string `json:"prefix"`
	LastFour   string `json:"lastFour"`
	Mode       string `json:"mode"`
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt"`
}

// GeneratedAPIKey is the one-time plaintext API key returned on creation.
type GeneratedAPIKey struct {
	PlainTextKey  string `json:"plainTextKey"`
	Prefix        string `json:"prefix"`
	LastFourChars string `json:"lastFourChars"`
	Mode          string `json:"mode"`
	CreatedAt     string `json:"createdAt"`
}

// List lists all API keys. GET /v1/api-keys
func (s *APIKeysService) List(ctx context.Context) ([]APIKey, error) {
	var out []APIKey
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/api-keys"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Create creates a new API key. POST /v1/api-keys?mode=LIVE
func (s *APIKeysService) Create(ctx context.Context, mode string) (*GeneratedAPIKey, error) {
	if mode == "" {
		mode = "LIVE"
	}
	var out GeneratedAPIKey
	err := s.client.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/api-keys",
		query:  url.Values{"mode": {mode}},
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
