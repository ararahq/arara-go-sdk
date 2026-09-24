package arara

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestShouldSendMessageWithGeneratedIdempotencyKey(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusAccepted, body: `{"id":"m1","status":"SENT","mode":"LIVE","sender":"s","receiver":"r","cost":0.35}`})
	resp, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "whatsapp:+5511999998888", TemplateName: "boas_vindas", TemplateVariables: []string{"Ana"}})
	if err != nil {
		t.Fatal(err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodPost, "/v1/messages")
	if !uuidV4Pattern.MatchString(req.Header.Get("Idempotency-Key")) {
		t.Fatalf("expected generated uuid key, got %q", req.Header.Get("Idempotency-Key"))
	}
	body := decodeBody(t, req.Body)
	if body["receiver"] != "whatsapp:+5511999998888" || body["templateName"] != "boas_vindas" {
		t.Fatalf("unexpected body %v", body)
	}
	if deref(resp.ID) != "m1" || resp.Cost == nil || *resp.Cost != 0.35 {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestShouldReuseSameIdempotencyKeyAcrossRetries(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), serverError(), fakeResponse{status: http.StatusAccepted, body: `{"id":"m1"}`})
	if _, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "5511999998888", Body: "oi"}); err != nil {
		t.Fatal(err)
	}
	if len(fs.requests) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(fs.requests))
	}
	first := fs.requests[0].Header.Get("Idempotency-Key")
	for i, r := range fs.requests {
		if r.Header.Get("Idempotency-Key") != first || first == "" {
			t.Fatalf("attempt %d used key %q, want %q", i, r.Header.Get("Idempotency-Key"), first)
		}
	}
}

func TestShouldUseCallerIdempotencyKey(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), fakeResponse{status: http.StatusAccepted, body: `{"id":"m1"}`})
	if _, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "+5511999998888", Body: "oi"}, SendOptions{IdempotencyKey: "order-42"}); err != nil {
		t.Fatal(err)
	}
	for _, r := range fs.requests {
		if r.Header.Get("Idempotency-Key") != "order-42" {
			t.Fatalf("expected caller key, got %q", r.Header.Get("Idempotency-Key"))
		}
	}
}

func TestShouldSendBatchWithIdempotencyKey(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusAccepted, body: `{"batchId":"b1","templateName":"t","total":2,"accepted":2,"totalCost":0.7,"messages":[{"id":"m1","receiver":"r","status":"QUEUED","cost":0.35}]}`})
	resp, err := c.Messages.SendBatch(bg, &BatchMessageRequest{TemplateName: "t", Messages: []BatchMessageItem{{Receiver: "5511999998888"}}})
	if err != nil {
		t.Fatal(err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodPost, "/v1/messages/batch")
	if req.Header.Get("Idempotency-Key") == "" {
		t.Fatal("batch must send Idempotency-Key")
	}
	if resp.BatchID != "b1" || resp.Accepted != 2 || len(resp.Messages) != 1 {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestShouldRejectBatchAboveLimitLocally(t *testing.T) {
	fs, c := newFakeServer(t)
	items := make([]BatchMessageItem, MaxBatchSize+1)
	if _, err := c.Messages.SendBatch(bg, &BatchMessageRequest{TemplateName: "t", Messages: items}); err == nil {
		t.Fatal("expected error")
	}
	if len(fs.requests) != 0 {
		t.Fatal("must not call the API")
	}
}

func TestShouldGetMessageByID(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"id":"abc","status":"FAILED","reason":"blocked"}`))
	resp, err := c.Messages.Get(bg, "abc")
	if err != nil {
		t.Fatal(err)
	}
	expectRoute(t, fs.only(), http.MethodGet, "/v1/messages/abc")
	if resp.Reason == nil || *resp.Reason != "blocked" {
		t.Fatalf("unexpected response %+v", resp)
	}
}

func TestShouldListMessagesByBatch(t *testing.T) {
	fs, c := newFakeServer(t, ok(`[{"id":"a"},{"id":"b"}]`))
	resp, err := c.Messages.ListByBatch(bg, "b1")
	if err != nil {
		t.Fatal(err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodGet, "/v1/messages")
	if !strings.Contains(req.Query, "batchId=b1") || len(resp) != 2 {
		t.Fatalf("unexpected query %q or response %+v", req.Query, resp)
	}
}

func TestShouldGenerateKeyWhenCallerKeyIsBlank(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), fakeResponse{status: http.StatusAccepted, body: `{"id":"m1"}`})
	if _, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "5511999998888", Body: "oi"}, SendOptions{IdempotencyKey: "   "}); err != nil {
		t.Fatal(err)
	}
	key := fs.requests[0].Header.Get("Idempotency-Key")
	if !uuidV4Pattern.MatchString(key) || fs.requests[1].Header.Get("Idempotency-Key") != key {
		t.Fatalf("blank key must be replaced by a stable uuid, got %q / %q", key, fs.requests[1].Header.Get("Idempotency-Key"))
	}
}

func TestShouldNotTreatBlankHeaderAsRetrySafe(t *testing.T) {
	if isRetrySafe(request{method: http.MethodPost, headers: map[string]string{"Idempotency-Key": "  "}}) {
		t.Fatal("blank Idempotency-Key must not make a POST retry-safe")
	}
	if !isRetrySafe(request{method: http.MethodPost, headers: map[string]string{"Idempotency-Key": "k"}}) {
		t.Fatal("POST with key must be retry-safe")
	}
}

func TestShouldValidateBatchLocally(t *testing.T) {
	fs, c := newFakeServer(t)
	cases := map[string]*BatchMessageRequest{
		"nil":            nil,
		"blank template": {TemplateName: "  ", Messages: []BatchMessageItem{{Receiver: "1"}}},
		"empty messages": {TemplateName: "t"},
	}
	for name, req := range cases {
		_, err := c.Messages.SendBatch(bg, req)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "INVALID_REQUEST" {
			t.Errorf("%s: expected INVALID_REQUEST, got %v", name, err)
		}
	}
	if len(fs.requests) != 0 {
		t.Fatal("invalid batches must not reach the API")
	}
}

func TestShouldDecodeNullMessageID(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusAccepted, body: `{"id":null,"status":"QUEUED"}`})
	resp, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "5511999998888", Body: "oi"})
	if err != nil || resp.ID != nil {
		t.Fatalf("expected nil id, got %+v %v", resp, err)
	}
}

func TestShouldSendTypedCharge(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusAccepted, body: `{"id":"m1"}`})
	expires := int64(1790000000)
	_, err := c.Messages.Send(bg, &SendMessageRequest{Receiver: "5511999998888", TemplateName: "cobranca", Charge: &ChargePayload{
		ReferenceID: "pedido-1",
		Items:       []ChargeItem{{Name: "Camisa", AmountCents: 4990, Quantity: 2}},
		Pix:         &PixPayment{Code: "000201", Key: "a@b.com", KeyType: "EMAIL", MerchantName: "Loja"},
		ExpiresAt:   &expires,
	}})
	if err != nil {
		t.Fatal(err)
	}
	charge, isMap := decodeBody(t, fs.only().Body)["charge"].(map[string]any)
	if !isMap || charge["referenceId"] != "pedido-1" || charge["pix"].(map[string]any)["keyType"] != "EMAIL" || charge["expiresAt"] != float64(expires) {
		t.Fatalf("unexpected charge %v", charge)
	}
	if _, has := charge["paymentLink"]; has {
		t.Fatal("unset payment methods must be omitted")
	}
}
