package arara

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ContactsService handles the /v1/contacts resource.
type ContactsService struct {
	client *Client
}

// ContactRequest is a contact to import.
type ContactRequest struct {
	Name       string         `json:"name"`
	Phone      string         `json:"phone"`
	Email      string         `json:"email,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// ContactPatchRequest is the payload for updating a contact.
type ContactPatchRequest struct {
	Name  string   `json:"name,omitempty"`
	Email string   `json:"email,omitempty"`
	Tags  []string `json:"tags,omitempty"`
}

// ContactResponse is a contact record.
type ContactResponse struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Phone            string         `json:"phone"`
	Email            *string        `json:"email"`
	Attributes       map[string]any `json:"attributes"`
	Tags             []string       `json:"tags"`
	CreatedAt        string         `json:"createdAt"`
	Lifecycle        string         `json:"lifecycle"`
	Source           string         `json:"source"`
	OutboundCount    int64          `json:"outboundCount"`
	InboundCount     int64          `json:"inboundCount"`
	FirstSeenAt      *string        `json:"firstSeenAt"`
	LastOutboundAt   *string        `json:"lastOutboundAt"`
	LastInboundAt    *string        `json:"lastInboundAt"`
	LastMessageAt    *string        `json:"lastMessageAt"`
	OptOutAt         *string        `json:"optOutAt"`
	LastTemplateName *string        `json:"lastTemplateName"`
}

// ContactsListResponse is a paginated list of contacts.
type ContactsListResponse struct {
	Contacts   []ContactResponse `json:"contacts"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Size       int               `json:"size"`
	TotalPages int               `json:"totalPages"`
}

// ContactsBatchError describes a single failed import row.
type ContactsBatchError struct {
	Index  int     `json:"index"`
	Phone  *string `json:"phone"`
	Reason string  `json:"reason"`
}

// ContactsBatchResponse is the result of a batch import.
type ContactsBatchResponse struct {
	ImportID string               `json:"importId"`
	Created  int                  `json:"created"`
	Updated  int                  `json:"updated"`
	Skipped  int                  `json:"skipped"`
	Errors   []ContactsBatchError `json:"errors"`
}

// ContactsStatsResponse holds contact lifecycle counters.
type ContactsStatsResponse struct {
	Total    int64 `json:"total"`
	NewCount int64 `json:"newCount"`
	Engaged  int64 `json:"engaged"`
	Silent   int64 `json:"silent"`
	Dormant  int64 `json:"dormant"`
	OptedOut int64 `json:"optedOut"`
}

// ContactsReactivationCandidate is a dormant contact eligible for reactivation.
type ContactsReactivationCandidate struct {
	Phone            string  `json:"phone"`
	Name             string  `json:"name"`
	LastMessageAt    *string `json:"lastMessageAt"`
	LastTemplateName *string `json:"lastTemplateName"`
}

// ContactsReactivationResponse lists reactivation candidates.
type ContactsReactivationResponse struct {
	Total      int64                           `json:"total"`
	Candidates []ContactsReactivationCandidate `json:"candidates"`
}

// ContactMessageItem is a single message in a contact's history.
type ContactMessageItem struct {
	ID           string  `json:"id"`
	Direction    string  `json:"direction"`
	Status       string  `json:"status"`
	TemplateName *string `json:"templateName"`
	Body         *string `json:"body"`
	CreatedAt    string  `json:"createdAt"`
}

// ContactMessagesResponse lists a contact's recent messages.
type ContactMessagesResponse struct {
	Phone    string               `json:"phone"`
	Total    int64                `json:"total"`
	Messages []ContactMessageItem `json:"messages"`
}

// ContactListParams are the optional filters for listing contacts.
type ContactListParams struct {
	Page      int
	Size      int
	Q         string
	Lifecycle string
}

// List lists contacts. GET /v1/contacts
func (s *ContactsService) List(ctx context.Context, params ContactListParams) (*ContactsListResponse, error) {
	q := url.Values{
		"page": {strconv.Itoa(params.Page)},
		"size": {strconv.Itoa(defaultSize(params.Size, 20))},
	}
	if params.Q != "" {
		q.Set("q", params.Q)
	}
	if params.Lifecycle != "" {
		q.Set("lifecycle", params.Lifecycle)
	}
	var out ContactsListResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ImportBatch imports a batch of contacts. POST /v1/contacts/batch
func (s *ContactsService) ImportBatch(ctx context.Context, contacts []ContactRequest) (*ContactsBatchResponse, error) {
	var out ContactsBatchResponse
	err := s.client.do(ctx, request{method: http.MethodPost, path: "/v1/contacts/batch", body: contacts}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Stats returns contact lifecycle stats. GET /v1/contacts/stats
func (s *ContactsService) Stats(ctx context.Context) (*ContactsStatsResponse, error) {
	var out ContactsStatsResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts/stats"}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ReactivationCandidates lists reactivation candidates. GET /v1/contacts/reactivation
func (s *ContactsService) ReactivationCandidates(ctx context.Context, limit int) (*ContactsReactivationResponse, error) {
	q := url.Values{"limit": {strconv.Itoa(defaultSize(limit, 100))}}
	var out ContactsReactivationResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts/reactivation", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTags lists distinct contact tags. GET /v1/contacts/tags
func (s *ContactsService) ListTags(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts/tags"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Get retrieves a contact by phone. GET /v1/contacts/{phone}
func (s *ContactsService) Get(ctx context.Context, phone string) (*ContactResponse, error) {
	var out ContactResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts/" + url.PathEscape(phone)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a contact by phone. PATCH /v1/contacts/{phone}
func (s *ContactsService) Update(ctx context.Context, phone string, patch *ContactPatchRequest) (*ContactResponse, error) {
	var out ContactResponse
	err := s.client.do(ctx, request{method: http.MethodPatch, path: "/v1/contacts/" + url.PathEscape(phone), body: patch}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Messages lists a contact's recent messages. GET /v1/contacts/{phone}/messages
func (s *ContactsService) Messages(ctx context.Context, phone string, limit int) (*ContactMessagesResponse, error) {
	q := url.Values{"limit": {strconv.Itoa(defaultSize(limit, 30))}}
	var out ContactMessagesResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/contacts/" + url.PathEscape(phone) + "/messages", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func defaultSize(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}
