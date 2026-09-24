package arara

import (
	"errors"
	"net/http"
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestShouldRejectEmptyAPIKey(t *testing.T) {
	if _, err := NewClient("  "); err == nil {
		t.Fatal("expected error for blank api key")
	}
}

func TestShouldGenerateValidUUIDv4(t *testing.T) {
	a, b := newUUIDv4(), newUUIDv4()
	if !uuidV4Pattern.MatchString(a) || a == b {
		t.Fatalf("invalid or repeated uuid: %s %s", a, b)
	}
}

func TestShouldRetryGetOnServerError(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), ok(`{"id":"m1","status":"SENT"}`))
	msg, err := c.Messages.Get(bg, "m1")
	if err != nil || deref(msg.ID) != "m1" {
		t.Fatalf("unexpected result %+v %v", msg, err)
	}
	if len(fs.requests) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(fs.requests))
	}
}

func TestShouldNotRetryPostWithoutIdempotencyKey(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), ok(`{"id":"t1"}`))
	_, err := c.Templates.Create(bg, &CreateTemplateRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(fs.requests) != 1 {
		t.Fatalf("POST without Idempotency-Key must not be retried, got %d attempts", len(fs.requests))
	}
}

func TestShouldNotRetryClientError(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusBadRequest, body: `{"error":{"code":"VALIDATION_ERROR","message":"bad","details":{"field":"receiver"}}}`})
	_, err := c.Messages.Get(bg, "m1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 || apiErr.Code != "VALIDATION_ERROR" || apiErr.Details["field"] != "receiver" {
		t.Fatalf("unexpected error %+v", apiErr)
	}
	if len(fs.requests) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(fs.requests))
	}
}

func TestShouldExposeRetryAfterOnRateLimit(t *testing.T) {
	limited := fakeResponse{status: http.StatusTooManyRequests, body: `{"error":{"code":"SEND_RATE_LIMITED","message":"slow","details":{}}}`, headers: map[string]string{"Retry-After": "0"}}
	_, c := newFakeServer(t, limited, limited, limited, limited)
	_, err := c.Messages.Get(bg, "m1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.RetryAfter == nil || *apiErr.RetryAfter != 0 || apiErr.StatusCode != 429 {
		t.Fatalf("expected 429 with RetryAfter, got %+v", err)
	}
}

func TestShouldMapPlanFeatureLocked(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusForbidden, body: `{"error":{"code":"PLAN_FEATURE_LOCKED","message":"Voo","details":{"feature":"brain","currentPlan":"DECOLAGEM","upgradeTo":"VOO"}}}`})
	_, err := c.Templates.Get(bg, "id")
	if !IsPlanFeatureLocked(err) || IsAuthError(err) {
		t.Fatalf("expected plan lock, got %v", err)
	}
	lock, _ := AsPlanFeatureLocked(err)
	if lock.Feature != "brain" || lock.CurrentPlan != "DECOLAGEM" || lock.UpgradeTo != "VOO" {
		t.Fatalf("unexpected lock %+v", lock)
	}
}

func TestShouldMapEmptyForbiddenToAmbiguousForbidden(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusForbidden, body: ``})
	_, err := c.Messages.Get(bg, "other-org-message")
	if !IsForbidden(err) || IsAuthError(err) || IsPlanFeatureLocked(err) || IsNotFound(err) {
		t.Fatalf("expected ambiguous FORBIDDEN, got %v", err)
	}
}

func TestShouldMapEmptyNotFoundToNotFound(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusNotFound, body: ``})
	_, err := c.Messages.Get(bg, "missing")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "NOT_FOUND" || !IsNotFound(err) || IsForbidden(err) {
		t.Fatalf("expected NOT_FOUND, got %v", err)
	}
}

func TestShouldMapUnauthorizedToAuthError(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusUnauthorized, body: `{"status":401}`})
	_, err := c.Auth.Me(bg)
	var apiErr *APIError
	if !IsAuthError(err) || !errors.As(err, &apiErr) || apiErr.Code != "AUTHENTICATION_ERROR" {
		t.Fatalf("expected auth error, got %v", err)
	}
}

func TestShouldReportNetworkErrorWithoutRetryingPost(t *testing.T) {
	c, _ := NewClient(testAPIKey, WithBaseURL("http://127.0.0.1:1"), WithMaxRetries(0))
	_, err := c.Templates.Create(bg, &CreateTemplateRequest{Name: "x"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != codeNetwork {
		t.Fatalf("expected network error, got %v", err)
	}
}

func TestShouldFailDecodeOnInvalidJSON(t *testing.T) {
	_, c := newFakeServer(t, ok(`not-json`))
	_, err := c.Messages.Get(bg, "m1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "DECODE_ERROR" {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestShouldExposeVersion(t *testing.T) {
	if Version != "1.0.0" {
		t.Fatalf("unexpected version %s", Version)
	}
}
