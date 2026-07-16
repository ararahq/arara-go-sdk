package arara

import "fmt"

// Error is the typed error returned for every non-2xx response and transport failure.
type Error struct {
	StatusCode int
	Code       string
	Message    string
	Details    map[string]any
	RetryAfter *int
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("arara: %s (code=%s, status=%d)", e.Message, e.Code, e.StatusCode)
	}
	return fmt.Sprintf("arara: %s (code=%s)", e.Message, e.Code)
}
