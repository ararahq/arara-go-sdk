package arara

import (
	"context"
	"net/http"
)

// AuthService handles the /auth/me resource. Requires an ADMIN API key.
type AuthService struct {
	client *Client
}

// User is the user that owns the API key.
type User struct {
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Role         *string `json:"role"`
	EmailPending bool    `json:"emailPending"`
}

// Me returns the user that owns the API key. GET /auth/me (ADMIN key only)
func (s *AuthService) Me(ctx context.Context) (*User, error) {
	var out User
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/auth/me"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
