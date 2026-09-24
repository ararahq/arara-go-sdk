package arara

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	codePlanFeatureLocked = "PLAN_FEATURE_LOCKED"
	codeAuthentication    = "AUTHENTICATION_ERROR"
	codeUnknown           = "UNKNOWN_ERROR"
	codeNetwork           = "NETWORK_ERROR"
	codeForbidden         = "FORBIDDEN"
	codeNotFound          = "NOT_FOUND"
	codeContextCanceled   = "CONTEXT_CANCELED"
)

// APIError is the typed error returned for every non-2xx response and transport failure.
// Use errors.As(err, &apiErr) to inspect it.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]any
	RetryAfter *int
	// Err is the underlying transport or context error, if any. Exposed through Unwrap, so
	// errors.Is(err, context.Canceled) works.
	Err error
}

// Unwrap returns the underlying transport or context error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("arara: %s (code=%s, status=%d)", e.Message, e.Code, e.StatusCode)
	}
	return fmt.Sprintf("arara: %s (code=%s)", e.Message, e.Code)
}

// PlanFeatureLock carries the details of a 403 PLAN_FEATURE_LOCKED response.
type PlanFeatureLock struct {
	Feature     string
	CurrentPlan string
	UpgradeTo   string
}

// IsPlanFeatureLocked reports whether err is a 403 PLAN_FEATURE_LOCKED from the API.
func IsPlanFeatureLocked(err error) bool {
	_, ok := AsPlanFeatureLocked(err)
	return ok
}

// AsPlanFeatureLocked extracts feature, currentPlan and upgradeTo from a PLAN_FEATURE_LOCKED error.
func AsPlanFeatureLocked(err error) (*PlanFeatureLock, bool) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return nil, false
	}
	if apiErr.StatusCode != http.StatusForbidden || apiErr.Code != codePlanFeatureLocked {
		return nil, false
	}
	return &PlanFeatureLock{
		Feature:     detailString(apiErr.Details, "feature"),
		CurrentPlan: detailString(apiErr.Details, "currentPlan"),
		UpgradeTo:   detailString(apiErr.Details, "upgradeTo"),
	}, true
}

// IsAuthError reports whether err is a 401 (the API key was rejected).
//
// A 403 without an error envelope carries Code "FORBIDDEN" and is deliberately NOT treated as an
// auth error: the API uses it both for key failures (invalid or expired key, missing permission,
// path outside the key allowlist) and for resources owned by another organization
// (GET /v1/messages/{id} answers an empty 403). Use IsForbidden to detect it.
func IsAuthError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusUnauthorized
}

// IsForbidden reports whether err is a 403 without an error code. The cause is ambiguous: an
// API key without permission for the route, or a resource that belongs to another organization.
func IsForbidden(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusForbidden && apiErr.Code == codeForbidden
}

// IsNotFound reports whether err is a 404.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusNotFound
}

func detailString(details map[string]any, key string) string {
	value, ok := details[key].(string)
	if !ok {
		return ""
	}
	return value
}
