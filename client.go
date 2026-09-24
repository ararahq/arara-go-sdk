// Package arara is the official Go SDK for the AraraHQ API.
package arara

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://api.ararahq.com"
	defaultTimeout    = 10 * time.Second
	defaultMaxRetries = 3

	baseRetryDelay   = 500 * time.Millisecond
	maxRetryDelay    = 30 * time.Second
	rateLimitStatus  = 429
	serverErrorFloor = 500
	uuidByteLength   = 16
)

// Client is the root AraraHQ API client. Access resources through its public fields.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	maxRetries int

	Messages      *MessagesService
	Templates     *TemplatesService
	Auth          *AuthService
	OptOuts       *OptOutsService
	Contacts      *ContactsService
	Conversations *ConversationsService
	Wallet        *WalletService
	Numbers       *NumbersService
	SmartLinks    *SmartLinksService
	Campaigns     *CampaignsService
}

// Option customizes a Client during construction.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient sets a custom HTTP client. The SDK never mutates it: when WithTimeout is also
// given, a shallow copy with that timeout is used instead.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the per-request timeout, regardless of option order.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithMaxRetries sets how many times 429/5xx/network failures are retried. Only GET requests
// and requests carrying an Idempotency-Key are ever retried.
func WithMaxRetries(n int) Option {
	return func(c *Client) { c.maxRetries = max(n, 0) }
}

// NewClient builds a client authenticated with the given API key.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, &APIError{Code: "INVALID_CONFIG", Message: "apiKey is required to instantiate the client"}
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		maxRetries: defaultMaxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.httpClient = resolveHTTPClient(c.httpClient, c.timeout)

	c.Messages = &MessagesService{client: c}
	c.Templates = &TemplatesService{client: c}
	c.Auth = &AuthService{client: c}
	c.OptOuts = &OptOutsService{client: c}
	c.Contacts = &ContactsService{client: c}
	c.Conversations = &ConversationsService{client: c}
	c.Wallet = &WalletService{client: c}
	c.Numbers = &NumbersService{client: c}
	c.SmartLinks = &SmartLinksService{client: c}
	c.Campaigns = &CampaignsService{client: c}
	return c, nil
}

func resolveHTTPClient(custom *http.Client, timeout time.Duration) *http.Client {
	if custom == nil {
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		return &http.Client{Timeout: timeout}
	}
	if timeout <= 0 {
		return custom
	}
	cp := *custom
	cp.Timeout = timeout
	return &cp
}

type request struct {
	method  string
	path    string
	query   url.Values
	body    any
	headers map[string]string
}

func (c *Client) do(ctx context.Context, req request, out any) error {
	var payload []byte
	if req.body != nil {
		encoded, err := json.Marshal(req.body)
		if err != nil {
			return &APIError{Code: "ENCODE_ERROR", Message: fmt.Sprintf("failed to encode request body: %v", err), Err: err}
		}
		payload = encoded
	}

	endpoint := c.baseURL + req.path
	if len(req.query) > 0 {
		endpoint += "?" + req.query.Encode()
	}

	retrySafe := isRetrySafe(req)
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := computeRetryDelay(attempt-1, retryAfterFromErr(lastErr))
			select {
			case <-ctx.Done():
				return contextError(ctx)
			case <-time.After(delay):
			}
		}

		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}
		httpReq, err := http.NewRequestWithContext(ctx, req.method, endpoint, bodyReader)
		if err != nil {
			return &APIError{Code: "REQUEST_ERROR", Message: err.Error(), Err: err}
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
		httpReq.Header.Set("Accept", "application/json")
		if payload != nil {
			httpReq.Header.Set("Content-Type", "application/json")
		}
		for k, v := range req.headers {
			httpReq.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			if ctx.Err() != nil {
				return contextError(ctx)
			}
			lastErr = &APIError{Code: codeNetwork, Message: err.Error(), Err: err}
			if !retrySafe {
				return lastErr
			}
			continue
		}

		apiErr := decodeResponse(resp, out)
		if apiErr == nil {
			return nil
		}
		lastErr = apiErr
		if !retrySafe || !isRetryable(apiErr) {
			return apiErr
		}
	}
	if lastErr == nil {
		return &APIError{Code: "NO_ATTEMPT", Message: "request was not attempted"}
	}
	return lastErr
}

func contextError(ctx context.Context) *APIError {
	return &APIError{Code: codeContextCanceled, Message: ctx.Err().Error(), Err: ctx.Err()}
}

func decodeResponse(resp *http.Response, out any) *APIError {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil || resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(raw)) == 0 {
			return nil
		}
		if err := json.Unmarshal(raw, out); err != nil {
			return &APIError{StatusCode: resp.StatusCode, Code: "DECODE_ERROR", Message: fmt.Sprintf("failed to decode response: %v", err)}
		}
		return nil
	}

	apiErr := parseErrorEnvelope(raw)
	apiErr.StatusCode = resp.StatusCode
	if apiErr.Code == "" {
		apiErr.Code = fallbackErrorCode(resp.StatusCode)
	}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("request failed with status %d", resp.StatusCode)
	}
	apiErr.RetryAfter = parseRetryAfter(resp.Header.Get("Retry-After"))
	return apiErr
}

func parseErrorEnvelope(raw []byte) *APIError {
	var envelope struct {
		Error struct {
			Code    string         `json:"code"`
			Message string         `json:"message"`
			Details map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return &APIError{}
	}
	return &APIError{
		Code:    envelope.Error.Code,
		Message: envelope.Error.Message,
		Details: envelope.Error.Details,
	}
}

func fallbackErrorCode(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return codeAuthentication
	case http.StatusForbidden:
		return codeForbidden
	case http.StatusNotFound:
		return codeNotFound
	default:
		return codeUnknown
	}
}

func isRetrySafe(req request) bool {
	if req.method == http.MethodGet || req.method == http.MethodHead {
		return true
	}
	return strings.TrimSpace(req.headers[idempotencyKeyHeader]) != ""
}

func isRetryable(e *APIError) bool {
	if e.Code == codeNetwork {
		return true
	}
	return e.StatusCode == rateLimitStatus || e.StatusCode >= serverErrorFloor
}

func retryAfterFromErr(err error) *int {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.RetryAfter
	}
	return nil
}

func computeRetryDelay(attempt int, retryAfterSeconds *int) time.Duration {
	if retryAfterSeconds != nil {
		d := time.Duration(*retryAfterSeconds) * time.Second
		if d > maxRetryDelay {
			return maxRetryDelay
		}
		return d
	}
	backoff := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt)))
	if backoff > maxRetryDelay {
		return maxRetryDelay
	}
	return backoff
}

func parseRetryAfter(header string) *int {
	header = strings.TrimSpace(header)
	if header == "" {
		return nil
	}
	if seconds, err := strconv.Atoi(header); err == nil && seconds >= 0 {
		return &seconds
	}
	if t, err := http.ParseTime(header); err == nil {
		seconds := int(math.Ceil(time.Until(t).Seconds()))
		if seconds < 0 {
			seconds = 0
		}
		return &seconds
	}
	return nil
}

func newUUIDv4() string {
	b := make([]byte, uuidByteLength)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("arara: crypto/rand unavailable: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
