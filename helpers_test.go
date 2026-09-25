package arara

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const testAPIKey = "ara_live_test"

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Header http.Header
	Body   []byte
}

type fakeResponse struct {
	status  int
	body    string
	headers map[string]string
}

type fakeServer struct {
	t         *testing.T
	mu        sync.Mutex
	requests  []recordedRequest
	responses []fakeResponse
}

func newFakeServer(t *testing.T, responses ...fakeResponse) (*fakeServer, *Client) {
	t.Helper()
	fs := &fakeServer{t: t, responses: responses}
	srv := httptest.NewServer(http.HandlerFunc(fs.handle))
	t.Cleanup(srv.Close)
	client, err := NewClient(testAPIKey, WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return fs, client
}

func (fs *fakeServer) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	fs.mu.Lock()
	idx := len(fs.requests)
	fs.requests = append(fs.requests, recordedRequest{
		Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Header: r.Header.Clone(), Body: body,
	})
	resp := fakeResponse{status: http.StatusOK, body: "{}"}
	if idx < len(fs.responses) {
		resp = fs.responses[idx]
	}
	fs.mu.Unlock()
	for k, v := range resp.headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.status)
	_, _ = io.WriteString(w, resp.body)
}

func (fs *fakeServer) only() recordedRequest {
	fs.t.Helper()
	if len(fs.requests) != 1 {
		fs.t.Fatalf("expected 1 request, got %d", len(fs.requests))
	}
	return fs.requests[0]
}

func ok(body string) fakeResponse { return fakeResponse{status: http.StatusOK, body: body} }

func serverError() fakeResponse {
	return fakeResponse{
		status:  http.StatusServiceUnavailable,
		body:    `{"error":{"code":"SEND_TEMPORARILY_UNAVAILABLE","message":"down","details":{}}}`,
		headers: map[string]string{"Retry-After": "0"},
	}
}

func expectRoute(t *testing.T, got recordedRequest, method, path string) {
	t.Helper()
	if got.Method != method || got.Path != path {
		t.Fatalf("expected %s %s, got %s %s", method, path, got.Method, got.Path)
	}
	if got.Header.Get("Authorization") != "Bearer "+testAPIKey {
		t.Fatalf("missing bearer auth, got %q", got.Header.Get("Authorization"))
	}
}

func decodeBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("invalid request body %q: %v", raw, err)
	}
	return out
}

var bg = context.Background()

func contextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
