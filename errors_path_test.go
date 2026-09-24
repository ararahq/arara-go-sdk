package arara

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestShouldReturnAPIErrorFromEveryResourceMethod(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":{"code":"INVALID_RECIPIENT","message":"bad","details":{}}}`))
	}))
	t.Cleanup(srv.Close)
	c, _ := NewClient(testAPIKey, WithBaseURL(srv.URL))
	calls := map[string]func() error{
		"Messages.Send": func() error { _, err := c.Messages.Send(bg, &SendMessageRequest{}); return err },
		"Messages.SendBatch": func() error {
			_, err := c.Messages.SendBatch(bg, &BatchMessageRequest{TemplateName: "t", Messages: []BatchMessageItem{{Receiver: "1"}}})
			return err
		},
		"Messages.Get":         func() error { _, err := c.Messages.Get(bg, "x"); return err },
		"Messages.ListByBatch": func() error { _, err := c.Messages.ListByBatch(bg, "x"); return err },
		"Templates.List":       func() error { _, err := c.Templates.List(bg, TemplateListParams{}); return err },
		"Templates.Create":     func() error { _, err := c.Templates.Create(bg, &CreateTemplateRequest{}); return err },
		"Templates.Get":        func() error { _, err := c.Templates.Get(bg, "x"); return err },
		"Templates.GetStatus":  func() error { _, err := c.Templates.GetStatus(bg, "x"); return err },
		"Templates.Delete":     func() error { return c.Templates.Delete(bg, "x") },
		"Templates.Analytics":  func() error { _, err := c.Templates.Analytics(bg, ""); return err },
		"Auth.Me":              func() error { _, err := c.Auth.Me(bg); return err },
		"SmartLinks.List":      func() error { _, err := c.SmartLinks.List(bg, PageParams{}); return err },
		"SmartLinks.Create": func() error {
			_, err := c.SmartLinks.Create(bg, &CreateWhatsAppSmartLinkRequest{})
			return err
		},
		"SmartLinks.Update": func() error {
			_, err := c.SmartLinks.Update(bg, "x", &UpdateWhatsAppSmartLinkRequest{})
			return err
		},
		"SmartLinks.Stats":       func() error { _, err := c.SmartLinks.Stats(bg, "x"); return err },
		"Campaigns.Create":       func() error { _, err := c.Campaigns.Create(bg, &CampaignRequest{}); return err },
		"Campaigns.List":         func() error { _, err := c.Campaigns.List(bg, 0, 0, ""); return err },
		"Campaigns.Estimate":     func() error { _, err := c.Campaigns.Estimate(bg, "t", 1); return err },
		"Campaigns.Get":          func() error { _, err := c.Campaigns.Get(bg, "x"); return err },
		"Contacts.List":          func() error { _, err := c.Contacts.List(bg, ContactListParams{}); return err },
		"Contacts.ImportBatch":   func() error { _, err := c.Contacts.ImportBatch(bg, nil); return err },
		"Contacts.Stats":         func() error { _, err := c.Contacts.Stats(bg); return err },
		"Contacts.Reactivation":  func() error { _, err := c.Contacts.ReactivationCandidates(bg, 1); return err },
		"Contacts.ListTags":      func() error { _, err := c.Contacts.ListTags(bg); return err },
		"Contacts.Get":           func() error { _, err := c.Contacts.Get(bg, "x"); return err },
		"Contacts.Update":        func() error { _, err := c.Contacts.Update(bg, "x", &ContactPatchRequest{}); return err },
		"Contacts.Messages":      func() error { _, err := c.Contacts.Messages(bg, "x", 1); return err },
		"Conversations.List":     func() error { _, err := c.Conversations.List(bg, ConversationListParams{}); return err },
		"Conversations.Stats":    func() error { _, err := c.Conversations.LeadStats(bg); return err },
		"Conversations.Messages": func() error { _, err := c.Conversations.Messages(bg, "x", 0, 0); return err },
		"Conversations.Reply": func() error {
			_, err := c.Conversations.Reply(bg, &ConversationReplyRequest{})
			return err
		},
		"Conversations.UpdateStatus": func() error { _, err := c.Conversations.UpdateStatus(bg, "x", "y"); return err },
		"Conversations.WindowStatus": func() error { _, err := c.Conversations.WindowStatus(bg, nil); return err },
		"Numbers.List":               func() error { _, err := c.Numbers.List(bg); return err },
		"Numbers.Update":             func() error { _, err := c.Numbers.Update(bg, "x", &UpdateNumberRequest{}); return err },
		"Numbers.Delete":             func() error { _, err := c.Numbers.Delete(bg, "x"); return err },
		"Numbers.Request":            func() error { _, err := c.Numbers.Request(bg, &RequestNumberRequest{}); return err },
		"Numbers.ListRequests":       func() error { _, err := c.Numbers.ListRequests(bg); return err },
		"Numbers.Sync":               func() error { _, err := c.Numbers.Sync(bg, "x"); return err },
		"Numbers.Warming":            func() error { _, err := c.Numbers.Warming(bg, "x"); return err },
		"Wallet.Transactions":        func() error { _, err := c.Wallet.Transactions(bg, 0, 0); return err },
		"Wallet.GetAutoRecharge":     func() error { _, err := c.Wallet.GetAutoRecharge(bg); return err },
		"Wallet.UpdateAutoRecharge": func() error {
			_, err := c.Wallet.UpdateAutoRecharge(bg, &UpdateAutoRechargeRequest{})
			return err
		},
	}
	for name, call := range calls {
		var apiErr *APIError
		if err := call(); !errors.As(err, &apiErr) || apiErr.Code != "INVALID_RECIPIENT" || apiErr.StatusCode != 422 {
			t.Errorf("%s: expected INVALID_RECIPIENT, got %v", name, err)
		}
	}
}

func TestShouldFormatErrorMessage(t *testing.T) {
	withStatus := (&APIError{StatusCode: 404, Code: "NOT_FOUND", Message: "gone"}).Error()
	withoutStatus := (&APIError{Code: codeNetwork, Message: "dial"}).Error()
	if !strings.Contains(withStatus, "status=404") || strings.Contains(withoutStatus, "status=") {
		t.Fatalf("unexpected messages %q %q", withStatus, withoutStatus)
	}
}

func TestShouldIgnoreNonAPIErrorsInHelpers(t *testing.T) {
	plain := errors.New("x")
	if IsPlanFeatureLocked(plain) || IsAuthError(plain) {
		t.Fatal("plain errors are not API errors")
	}
	if detailString(map[string]any{"feature": 1}, "feature") != "" {
		t.Fatal("non-string detail must be empty")
	}
	if retryAfterFromErr(plain) != nil {
		t.Fatal("plain error has no retry-after")
	}
}

func TestShouldComputeRetryDelay(t *testing.T) {
	big, small := 120, 2
	cases := []struct {
		attempt    int
		retryAfter *int
		want       time.Duration
	}{
		{0, nil, baseRetryDelay},
		{2, nil, 4 * baseRetryDelay},
		{20, nil, maxRetryDelay},
		{0, &small, 2 * time.Second},
		{0, &big, maxRetryDelay},
	}
	for _, tc := range cases {
		if got := computeRetryDelay(tc.attempt, tc.retryAfter); got != tc.want {
			t.Errorf("attempt %d: got %v want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestShouldParseRetryAfterHeader(t *testing.T) {
	if parseRetryAfter("") != nil || parseRetryAfter("soon") != nil {
		t.Fatal("invalid headers must be nil")
	}
	if v := parseRetryAfter("7"); v == nil || *v != 7 {
		t.Fatalf("unexpected %v", v)
	}
	past := time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)
	if v := parseRetryAfter(past); v == nil || *v != 0 {
		t.Fatalf("past date must clamp to 0, got %v", v)
	}
	future := time.Now().Add(time.Minute).UTC().Format(http.TimeFormat)
	if v := parseRetryAfter(future); v == nil || *v <= 0 {
		t.Fatalf("future date must be positive, got %v", v)
	}
}

func TestShouldApplyTimeoutWithoutMutatingCallerClient(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}
	c, _ := NewClient(testAPIKey, WithHTTPClient(custom), WithTimeout(3*time.Second), WithMaxRetries(1), WithBaseURL("http://x/"))
	if custom.Timeout != time.Second {
		t.Fatalf("caller client was mutated: %v", custom.Timeout)
	}
	if c.httpClient == custom || c.httpClient.Timeout != 3*time.Second || c.maxRetries != 1 || c.baseURL != "http://x" {
		t.Fatalf("options not applied: %+v", c)
	}
}

func TestShouldApplyTimeoutInAnyOrder(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}
	c, _ := NewClient(testAPIKey, WithTimeout(3*time.Second), WithHTTPClient(custom))
	if c.httpClient.Timeout != 3*time.Second || custom.Timeout != time.Second {
		t.Fatalf("timeout before custom client was lost: %v", c.httpClient.Timeout)
	}
}

func TestShouldKeepCustomClientWithoutTimeoutOption(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}
	c, _ := NewClient(testAPIKey, WithHTTPClient(custom))
	if c.httpClient != custom {
		t.Fatal("custom client must be used as-is")
	}
}

func TestShouldUseTimeoutOnDefaultClient(t *testing.T) {
	c, _ := NewClient(testAPIKey, WithTimeout(4*time.Second))
	d, _ := NewClient(testAPIKey)
	if c.httpClient.Timeout != 4*time.Second || d.httpClient.Timeout != defaultTimeout {
		t.Fatalf("unexpected timeouts %v %v", c.httpClient.Timeout, d.httpClient.Timeout)
	}
}

func TestShouldClampNegativeRetriesAndStillAttemptOnce(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"id":"m1"}`))
	c.maxRetries = 0
	WithMaxRetries(-1)(c)
	if c.maxRetries != 0 {
		t.Fatalf("expected clamp to 0, got %d", c.maxRetries)
	}
	msg, err := c.Messages.Get(bg, "m1")
	if err != nil || msg == nil || len(fs.requests) != 1 {
		t.Fatalf("expected one attempt, got %+v %v", msg, err)
	}
}

func TestShouldErrorWhenNoAttemptIsMade(t *testing.T) {
	c, _ := NewClient(testAPIKey)
	c.maxRetries = -1
	_, err := c.Messages.Get(bg, "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "NO_ATTEMPT" {
		t.Fatalf("expected NO_ATTEMPT, got %v", err)
	}
}

func TestShouldUnwrapContextCancellationWithoutRetrying(t *testing.T) {
	fs, c := newFakeServer(t)
	ctx, cancel := contextWithCancel()
	cancel()
	_, err := c.Messages.Get(ctx, "x")
	var apiErr *APIError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &apiErr) || apiErr.Code != "CONTEXT_CANCELED" {
		t.Fatalf("expected unwrappable cancellation, got %v", err)
	}
	if len(fs.requests) != 0 {
		t.Fatalf("canceled request must not be retried, got %d", len(fs.requests))
	}
}

func TestShouldStopRetryingWhenContextCanceled(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusServiceUnavailable, body: `{}`, headers: map[string]string{"Retry-After": "30"}})
	ctx, cancel := contextWithCancel()
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	_, err := c.Messages.Get(ctx, "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "CONTEXT_CANCELED" || !errors.Is(err, context.Canceled) || len(fs.requests) != 1 {
		t.Fatalf("expected cancellation after 1 attempt, got %v (%d)", err, len(fs.requests))
	}
}

func TestShouldRetryGetOnNetworkError(t *testing.T) {
	c, _ := NewClient(testAPIKey, WithBaseURL("http://127.0.0.1:1"), WithMaxRetries(1))
	start := time.Now()
	_, err := c.Messages.Get(bg, "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != codeNetwork || time.Since(start) < baseRetryDelay {
		t.Fatalf("expected retried network error, got %v", err)
	}
}

func TestShouldFailEncodingUnsupportedBody(t *testing.T) {
	_, c := newFakeServer(t)
	_, err := c.Templates.Create(bg, &CreateTemplateRequest{Samples: map[string]any{"x": make(chan int)}})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "ENCODE_ERROR" {
		t.Fatalf("expected encode error, got %v", err)
	}
}

func TestShouldFailOnInvalidRequestURL(t *testing.T) {
	c, _ := NewClient(testAPIKey, WithBaseURL("http://bad host"))
	_, err := c.Messages.Get(bg, "x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "REQUEST_ERROR" {
		t.Fatalf("expected request error, got %v", err)
	}
}
