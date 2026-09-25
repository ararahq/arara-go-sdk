package arara

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

const (
	campaignsBase            = "/v1/campaigns"
	defaultCampaignsPageSize = 20
)

// CampaignListParams are the optional filters for listing campaigns.
type CampaignListParams struct {
	Page   int
	Size   int
	Status string
}

// CampaignsService handles the /v1/campaigns resource.
type CampaignsService struct {
	client *Client
}

// CampaignContactRequest is a single campaign recipient.
type CampaignContactRequest struct {
	To        string   `json:"to"`
	Variables []string `json:"variables,omitempty"`
}

// CampaignAbConfig is the optional A/B test configuration. Omitted fields use server defaults.
type CampaignAbConfig struct {
	VariantBTemplateName  string `json:"variantBTemplateName"`
	Metric                string `json:"metric,omitempty"`
	SamplePct             *int   `json:"samplePct,omitempty"`
	SplitPct              *int   `json:"splitPct,omitempty"`
	DecisionWindowMinutes *int   `json:"decisionWindowMinutes,omitempty"`
	Autopilot             *bool  `json:"autopilot,omitempty"`
}

// CampaignRequest is the payload for creating a campaign.
type CampaignRequest struct {
	Name         string                   `json:"name"`
	TemplateName string                   `json:"templateName"`
	Sender       string                   `json:"sender,omitempty"`
	Contacts     []CampaignContactRequest `json:"contacts"`
	AbTest       *CampaignAbConfig        `json:"abTest,omitempty"`
	// ScheduledAt is an ISO-8601 instant (e.g. 2026-10-01T12:00:00Z). Empty sends immediately.
	ScheduledAt string `json:"scheduledAt,omitempty"`
}

// CampaignResponse is the response for a created campaign.
type CampaignResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Status        string  `json:"status"`
	TotalMessages int     `json:"totalMessages"`
	TotalCost     float64 `json:"totalCost"`
}

// CampaignListItem is a summarized campaign in a listing.
type CampaignListItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	TemplateName   string  `json:"templateName"`
	TotalMessages  int     `json:"totalMessages"`
	SentCount      int     `json:"sentCount"`
	DeliveredCount int     `json:"deliveredCount"`
	ReadCount      int     `json:"readCount"`
	TotalCost      float64 `json:"totalCost"`
	CreatedAt      *string `json:"createdAt"`
}

// CampaignListResponse is a paginated list of campaigns.
type CampaignListResponse struct {
	Content       []CampaignListItem `json:"content"`
	TotalPages    int                `json:"totalPages"`
	TotalElements int64              `json:"totalElements"`
}

// CampaignDetailResponse is the detailed view of a campaign.
type CampaignDetailResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	TemplateName   string  `json:"templateName"`
	TemplateBody   *string `json:"templateBody"`
	TotalMessages  int     `json:"totalMessages"`
	SentCount      int     `json:"sentCount"`
	DeliveredCount int     `json:"deliveredCount"`
	ReadCount      int     `json:"readCount"`
	ClickedCount   int     `json:"clickedCount"`
	ConvertedCount int     `json:"convertedCount"`
	ConvertedValue float64 `json:"convertedValue"`
	ReplyCount     int     `json:"replyCount"`
	HoldoutCount   int     `json:"holdoutCount"`
	BlockedCount   int     `json:"blockedCount"`
	RefundCount    int     `json:"refundCount"`
	RefundValue    float64 `json:"refundValue"`
	TotalCost      float64 `json:"totalCost"`
	ScheduledAt    *string `json:"scheduledAt"`
	StartedAt      *string `json:"startedAt"`
	FinishedAt     *string `json:"finishedAt"`
	CreatedAt      *string `json:"createdAt"`
}

// CampaignEstimateResponse is the cost estimate for a campaign.
type CampaignEstimateResponse struct {
	TemplateCategory string  `json:"templateCategory"`
	RecipientCount   int     `json:"recipientCount"`
	TemplateCost     float64 `json:"templateCost"`
	AraraFee         float64 `json:"araraFee"`
	UnitPrice        float64 `json:"unitPrice"`
	TotalCost        float64 `json:"totalCost"`
}

// CampaignCreateOptions carries per-call options for creating a campaign.
type CampaignCreateOptions struct {
	// IdempotencyKey is sent as the Idempotency-Key header. A UUID v4 is generated when empty.
	IdempotencyKey string
}

// Create creates a campaign. POST /v1/campaigns
func (s *CampaignsService) Create(ctx context.Context, req *CampaignRequest, opts ...CampaignCreateOptions) (*CampaignResponse, error) {
	sendOpts := make([]SendOptions, 0, 1)
	if len(opts) > 0 {
		sendOpts = append(sendOpts, SendOptions{IdempotencyKey: opts[0].IdempotencyKey})
	}
	var out CampaignResponse
	err := s.client.do(ctx, request{
		method:  http.MethodPost,
		path:    campaignsBase,
		body:    req,
		headers: idempotencyHeaders(sendOpts),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List lists campaigns. GET /v1/campaigns
func (s *CampaignsService) List(ctx context.Context, params CampaignListParams) (*CampaignListResponse, error) {
	q := PageParams{Page: params.Page, Size: params.Size}.values(defaultCampaignsPageSize)
	if params.Status != "" {
		q.Set("status", params.Status)
	}
	var out CampaignListResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: campaignsBase, query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Estimate estimates campaign cost. GET /v1/campaigns/estimate
func (s *CampaignsService) Estimate(ctx context.Context, templateName string, count int) (*CampaignEstimateResponse, error) {
	q := url.Values{
		"templateName": {templateName},
		"count":        {strconv.Itoa(count)},
	}
	var out CampaignEstimateResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: campaignsBase + "/estimate", query: q}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a campaign by id. GET /v1/campaigns/{id}
func (s *CampaignsService) Get(ctx context.Context, id string) (*CampaignDetailResponse, error) {
	var out CampaignDetailResponse
	err := s.client.do(ctx, request{method: http.MethodGet, path: campaignsBase + "/" + url.PathEscape(id)}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a campaign. POST /v1/campaigns/{id}/cancel
func (s *CampaignsService) Cancel(ctx context.Context, id string) error {
	return s.client.do(ctx, request{method: http.MethodPost, path: campaignsBase + "/" + url.PathEscape(id) + "/cancel"}, nil)
}
