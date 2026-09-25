package arara

import (
	"net/http"
	"strings"
	"testing"
)

func TestShouldReturnAuthenticatedUser(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"name":"Ana","email":"ana@x.com","role":"ADMIN","emailPending":false}`))
	user, err := c.Auth.Me(bg)
	if err != nil || user.Email != "ana@x.com" || user.Role == nil || *user.Role != "ADMIN" {
		t.Fatalf("unexpected %+v %v", user, err)
	}
	expectRoute(t, fs.only(), http.MethodGet, "/auth/me")
}

func TestShouldListSmartLinksPaginated(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"data":[{"id":"s1","name":"n","clicks":4}],"pagination":{"page":0,"size":50,"totalElements":1,"totalPages":1}}`))
	page, err := c.SmartLinks.List(bg, PageParams{})
	if err != nil || len(page.Data) != 1 || page.Data[0].Clicks != 4 || page.Pagination.TotalPages != 1 {
		t.Fatalf("unexpected %+v %v", page, err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodGet, "/v1/smart-links/whatsapp")
	if !strings.Contains(req.Query, "size=50") {
		t.Fatalf("unexpected query %q", req.Query)
	}
}

func TestShouldCreateCampaignWithStableIdempotencyKey(t *testing.T) {
	fs, c := newFakeServer(t, serverError(), ok(`{"id":"c1","status":"SCHEDULED","totalMessages":1}`))
	resp, err := c.Campaigns.Create(bg, &CampaignRequest{Name: "n", TemplateName: "t", Contacts: []CampaignContactRequest{{To: "5511999998888"}}, ScheduledAt: "2026-10-01T12:00:00Z"})
	if err != nil || resp.ID != "c1" {
		t.Fatalf("unexpected %+v %v", resp, err)
	}
	if len(fs.requests) != 2 {
		t.Fatalf("expected retry, got %d attempts", len(fs.requests))
	}
	key := fs.requests[0].Header.Get("Idempotency-Key")
	if !uuidV4Pattern.MatchString(key) || fs.requests[1].Header.Get("Idempotency-Key") != key {
		t.Fatalf("keys differ: %q %q", key, fs.requests[1].Header.Get("Idempotency-Key"))
	}
	body := decodeBody(t, fs.requests[0].Body)
	if body["scheduledAt"] != "2026-10-01T12:00:00Z" {
		t.Fatalf("scheduledAt missing: %v", body)
	}
}

func TestShouldCreateCampaignWithCallerKey(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"id":"c1"}`))
	if _, err := c.Campaigns.Create(bg, &CampaignRequest{Name: "n"}, CampaignCreateOptions{IdempotencyKey: "camp-1"}); err != nil {
		t.Fatal(err)
	}
	if fs.only().Header.Get("Idempotency-Key") != "camp-1" {
		t.Fatal("expected caller key")
	}
}

func TestShouldCallCampaignReadEndpoints(t *testing.T) {
	fs, c := newFakeServer(t,
		ok(`{"content":[{"id":"c1"}],"totalPages":1,"totalElements":1}`),
		ok(`{"id":"c1","name":"n"}`),
		ok(`{"totalCost":1.5,"recipientCount":3}`),
		fakeResponse{status: http.StatusOK},
	)
	list, err := c.Campaigns.List(bg, CampaignListParams{Status: "RUNNING"})
	if err != nil || len(list.Content) != 1 || list.TotalElements != 1 {
		t.Fatalf("list %+v %v", list, err)
	}
	if _, err := c.Campaigns.Get(bg, "c1"); err != nil {
		t.Fatal(err)
	}
	est, err := c.Campaigns.Estimate(bg, "t", 3)
	if err != nil || est.RecipientCount != 3 {
		t.Fatalf("estimate %+v %v", est, err)
	}
	if err := c.Campaigns.Cancel(bg, "c1"); err != nil {
		t.Fatal(err)
	}
	expectRoute(t, fs.requests[0], http.MethodGet, "/v1/campaigns")
	if !strings.Contains(fs.requests[0].Query, "status=RUNNING") || !strings.Contains(fs.requests[0].Query, "size=20") {
		t.Fatalf("unexpected query %q", fs.requests[0].Query)
	}
	expectRoute(t, fs.requests[1], http.MethodGet, "/v1/campaigns/c1")
	expectRoute(t, fs.requests[2], http.MethodGet, "/v1/campaigns/estimate")
	if !strings.Contains(fs.requests[2].Query, "count=3") {
		t.Fatalf("unexpected query %q", fs.requests[2].Query)
	}
	expectRoute(t, fs.requests[3], http.MethodPost, "/v1/campaigns/c1/cancel")
}

func TestShouldCallOptOutEndpoints(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"items":[]}`), ok(`{"optedOut":true}`), ok(`{"ok":true}`), ok(`{"ok":true}`))
	if _, err := c.OptOuts.List(bg); err != nil {
		t.Fatal(err)
	}
	if _, err := c.OptOuts.Get(bg, "5511999998888"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.OptOuts.Create(bg, &OptOutRequest{Phone: "5511999998888", Reason: "pediu"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.OptOuts.Delete(bg, "5511999998888"); err != nil {
		t.Fatal(err)
	}
	expectRoute(t, fs.requests[0], http.MethodGet, "/v1/opt-outs")
	expectRoute(t, fs.requests[1], http.MethodGet, "/v1/opt-outs/5511999998888")
	expectRoute(t, fs.requests[2], http.MethodPost, "/v1/opt-outs")
	if decodeBody(t, fs.requests[2].Body)["phone"] != "5511999998888" {
		t.Fatalf("unexpected body %s", fs.requests[2].Body)
	}
	expectRoute(t, fs.requests[3], http.MethodDelete, "/v1/opt-outs/5511999998888")
}

func TestShouldReturnErrorFromOptOuts(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusForbidden})
	if _, err := c.OptOuts.List(bg); !IsForbidden(err) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestShouldCallContactEndpoints(t *testing.T) {
	fs, c := newFakeServer(t,
		ok(`{"contacts":[{"id":"1","phone":"p"}],"total":1,"page":0,"size":20,"totalPages":1}`),
		ok(`{"created":1}`),
		ok(`{"total":5}`),
		ok(`{"total":1,"candidates":[]}`),
		ok(`{"tags":["vip"]}`),
		ok(`{"id":"1","phone":"5511"}`),
		ok(`{"id":"1","name":"Ana"}`),
		ok(`{"phone":"5511","total":0,"messages":[]}`),
	)
	list, err := c.Contacts.List(bg, ContactListParams{Q: "ana", Lifecycle: "ENGAGED"})
	if err != nil || list.Total != 1 {
		t.Fatalf("list %+v %v", list, err)
	}
	if res, err := c.Contacts.ImportBatch(bg, []ContactRequest{{Name: "Ana", Phone: "5511"}}); err != nil || res.Created != 1 {
		t.Fatalf("batch %+v %v", res, err)
	}
	if res, err := c.Contacts.Stats(bg); err != nil || res.Total != 5 {
		t.Fatalf("stats %+v %v", res, err)
	}
	if _, err := c.Contacts.ReactivationCandidates(bg, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Contacts.ListTags(bg); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Contacts.Get(bg, "5511"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Contacts.Update(bg, "5511", &ContactPatchRequest{Name: "Ana"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Contacts.Messages(bg, "5511", 0); err != nil {
		t.Fatal(err)
	}
	routes := [][2]string{
		{http.MethodGet, "/v1/contacts"}, {http.MethodPost, "/v1/contacts/batch"}, {http.MethodGet, "/v1/contacts/stats"},
		{http.MethodGet, "/v1/contacts/reactivation"}, {http.MethodGet, "/v1/contacts/tags"}, {http.MethodGet, "/v1/contacts/5511"},
		{http.MethodPatch, "/v1/contacts/5511"}, {http.MethodGet, "/v1/contacts/5511/messages"},
	}
	for i, r := range routes {
		expectRoute(t, fs.requests[i], r[0], r[1])
	}
	if !strings.Contains(fs.requests[0].Query, "q=ana") || !strings.Contains(fs.requests[0].Query, "lifecycle=ENGAGED") {
		t.Fatalf("unexpected query %q", fs.requests[0].Query)
	}
}

func TestShouldCallConversationEndpoints(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{}`), ok(`{}`), ok(`{}`), ok(`{}`), ok(`{}`), ok(`{}`))
	calls := []func() error{
		func() error {
			_, err := c.Conversations.List(bg, ConversationListParams{Status: "OPEN", LeadStatus: "HOT"})
			return err
		},
		func() error { _, err := c.Conversations.LeadStats(bg); return err },
		func() error { _, err := c.Conversations.Messages(bg, "cv1", 0, 0); return err },
		func() error {
			_, err := c.Conversations.Reply(bg, &ConversationReplyRequest{ConversationID: "cv1", Body: "oi"})
			return err
		},
		func() error { _, err := c.Conversations.UpdateStatus(bg, "cv1", "CLOSED"); return err },
		func() error { _, err := c.Conversations.WindowStatus(bg, []string{"5511"}); return err },
	}
	for i, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	routes := [][2]string{
		{http.MethodGet, "/v1/conversations"}, {http.MethodGet, "/v1/conversations/lead-stats"},
		{http.MethodGet, "/v1/conversations/cv1/messages"}, {http.MethodPost, "/v1/conversations/reply"},
		{http.MethodPatch, "/v1/conversations/cv1/status"}, {http.MethodPost, "/v1/conversations/window-status"},
	}
	for i, r := range routes {
		expectRoute(t, fs.requests[i], r[0], r[1])
	}
}

func TestShouldCallNumberEndpoints(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"numbers":[{"id":"n1"}],"slot":{"used":1,"max":2}}`), ok(`{}`), ok(`{}`), ok(`{}`), ok(`[{"id":"r1"}]`), ok(`{}`), ok(`{}`))
	list, err := c.Numbers.List(bg)
	if err != nil || len(list.Numbers) != 1 || list.Slot.Max != 2 {
		t.Fatalf("list %+v %v", list, err)
	}
	isDefault := true
	calls := []func() error{
		func() error {
			_, err := c.Numbers.Update(bg, "n1", &UpdateNumberRequest{IsDefault: &isDefault})
			return err
		},
		func() error { _, err := c.Numbers.Delete(bg, "n1"); return err },
		func() error { _, err := c.Numbers.Request(bg, &RequestNumberRequest{AreaCode: "11"}); return err },
		func() error { _, err := c.Numbers.ListRequests(bg); return err },
		func() error { _, err := c.Numbers.Sync(bg, "n1"); return err },
		func() error { _, err := c.Numbers.Warming(bg, "n1"); return err },
	}
	for i, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	base := "/v1/organizations/me/numbers"
	routes := [][2]string{
		{http.MethodGet, base}, {http.MethodPatch, base + "/n1"}, {http.MethodDelete, base + "/n1"},
		{http.MethodPost, base + "/request"}, {http.MethodGet, base + "/requests"},
		{http.MethodPost, base + "/n1/sync"}, {http.MethodGet, base + "/n1/warming"},
	}
	for i, r := range routes {
		expectRoute(t, fs.requests[i], r[0], r[1])
	}
}

func TestShouldCallWalletEndpoints(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"content":[{"id":"t1","amount":10}],"totalElements":1,"totalPages":1}`), ok(`{"enabled":true}`), ok(`{"enabled":false}`))
	page, err := c.Wallet.Transactions(bg, 0, 0)
	if err != nil || len(page.Content) != 1 || page.TotalElements != 1 {
		t.Fatalf("transactions %+v %v", page, err)
	}
	if s, err := c.Wallet.GetAutoRecharge(bg); err != nil || !s.Enabled {
		t.Fatalf("get %+v %v", s, err)
	}
	disabled := false
	if s, err := c.Wallet.UpdateAutoRecharge(bg, &UpdateAutoRechargeRequest{Enabled: &disabled}); err != nil || s.Enabled {
		t.Fatalf("update %+v %v", s, err)
	}
	expectRoute(t, fs.requests[0], http.MethodGet, "/v1/wallet/transactions")
	expectRoute(t, fs.requests[1], http.MethodGet, "/v1/wallet/auto-recharge")
	expectRoute(t, fs.requests[2], http.MethodPatch, "/v1/wallet/auto-recharge")
}

func TestShouldCallSmartLinkWriteEndpoints(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"id":"s1"}`), ok(`{"id":"s1","name":"b"}`), ok(`{"clicks":2}`))
	if _, err := c.SmartLinks.Create(bg, &CreateWhatsAppSmartLinkRequest{Name: "a", PhoneNumber: "5511"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SmartLinks.Update(bg, "s1", &UpdateWhatsAppSmartLinkRequest{Name: "b"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SmartLinks.Stats(bg, "s1"); err != nil {
		t.Fatal(err)
	}
	expectRoute(t, fs.requests[0], http.MethodPost, "/v1/smart-links/whatsapp")
	expectRoute(t, fs.requests[1], http.MethodPut, "/v1/smart-links/whatsapp/s1")
	expectRoute(t, fs.requests[2], http.MethodGet, "/v1/smart-links/whatsapp/s1/stats")
}
