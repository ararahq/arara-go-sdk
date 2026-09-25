package arara

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ConversationsService handles the /v1/conversations resource.
type ConversationsService struct {
	client *Client
}

// ConversationReplyRequest replies within an open 24h window.
type ConversationReplyRequest struct {
	ConversationID string `json:"conversationId"`
	Body           string `json:"body"`
}

// ConversationListParams are the optional filters for listing conversations.
type ConversationListParams struct {
	Status     string
	LeadStatus string
	Page       int
	Size       int
}

// List lists conversations. GET /v1/conversations
func (s *ConversationsService) List(ctx context.Context, params ConversationListParams) (map[string]any, error) {
	q := url.Values{
		"page": {strconv.Itoa(params.Page)},
		"size": {strconv.Itoa(defaultSize(params.Size, 20))},
	}
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	if params.LeadStatus != "" {
		q.Set("leadStatus", params.LeadStatus)
	}
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/conversations", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LeadStats returns lead status stats. GET /v1/conversations/lead-stats
func (s *ConversationsService) LeadStats(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/conversations/lead-stats"}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Messages lists messages in a conversation. GET /v1/conversations/{conversationId}/messages
func (s *ConversationsService) Messages(ctx context.Context, conversationID string, page, size int) (map[string]any, error) {
	q := url.Values{
		"page": {strconv.Itoa(page)},
		"size": {strconv.Itoa(defaultSize(size, 50))},
	}
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodGet, path: "/v1/conversations/" + url.PathEscape(conversationID) + "/messages", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Reply replies within an open conversation window. POST /v1/conversations/reply
func (s *ConversationsService) Reply(ctx context.Context, req *ConversationReplyRequest) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{method: http.MethodPost, path: "/v1/conversations/reply", body: req}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateStatus updates a conversation status. PATCH /v1/conversations/{conversationId}/status
func (s *ConversationsService) UpdateStatus(ctx context.Context, conversationID, status string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{
		method: http.MethodPatch,
		path:   "/v1/conversations/" + url.PathEscape(conversationID) + "/status",
		body:   map[string]string{"status": status},
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WindowStatus checks the 24h window status for phones. POST /v1/conversations/window-status
func (s *ConversationsService) WindowStatus(ctx context.Context, phones []string) (map[string]any, error) {
	var out map[string]any
	err := s.client.do(ctx, request{
		method: http.MethodPost,
		path:   "/v1/conversations/window-status",
		body:   map[string][]string{"phones": phones},
	}, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
