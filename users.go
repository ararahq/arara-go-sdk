package arara

import (
	"context"
	"net/http"
)

// UsersService handles the /users resource.
type UsersService struct {
	client *Client
}

// User is the authenticated user.
type User struct {
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	PhoneNumber            string `json:"phoneNumber,omitempty"`
	NeedsInitialOnboarding *bool  `json:"needsInitialOnboarding,omitempty"`
}

// UpdateUserRequest is the payload for updating the authenticated user.
type UpdateUserRequest struct {
	Name        string `json:"name,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

// GetMe returns the authenticated user. GET /users/me
func (s *UsersService) GetMe(ctx context.Context) (*User, error) {
	var out User
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/users/me"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates the authenticated user. PATCH /users/me
func (s *UsersService) Update(ctx context.Context, req *UpdateUserRequest) (*User, error) {
	var out User
	err := s.client.do(ctx, request{method: http.MethodPatch, path: "/users/me", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
