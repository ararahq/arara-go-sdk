package arara

import (
	"net/http"
	"strings"
	"testing"
)

const templateID = "0b7f5c9e-3f5a-4d57-9a55-2b0c7f1d9e10"

func TestShouldListTemplatesPaginated(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"data":[{"id":"`+templateID+`","name":"boas_vindas","providerStatus":"APPROVED"}],"pagination":{"page":1,"size":10,"totalElements":11,"totalPages":2}}`))
	page, err := c.Templates.List(bg, TemplateListParams{Name: "boas_vindas", Status: "APPROVED", Page: 1, Size: 10})
	if err != nil {
		t.Fatal(err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodGet, "/v1/templates")
	for _, want := range []string{"page=1", "size=10", "name=boas_vindas", "status=APPROVED"} {
		if !strings.Contains(req.Query, want) {
			t.Fatalf("query %q missing %s", req.Query, want)
		}
	}
	if len(page.Data) != 1 || page.Data[0].ID != templateID || page.Pagination.TotalElements != 11 || page.Pagination.TotalPages != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestShouldUseDefaultTemplatePageSize(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"data":[],"pagination":{"page":0,"size":50,"totalElements":0,"totalPages":0}}`))
	if _, err := c.Templates.List(bg, TemplateListParams{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fs.only().Query, "size=50") {
		t.Fatalf("unexpected query %q", fs.only().Query)
	}
}

func TestShouldFindTemplateByName(t *testing.T) {
	_, c := newFakeServer(t, ok(`{"data":[{"id":"x","name":"boas_vindas_b"},{"id":"y","name":"boas_vindas"}],"pagination":{}}`))
	tpl, err := c.Templates.FindByName(bg, "boas_vindas")
	if err != nil || tpl.ID != "y" {
		t.Fatalf("unexpected %+v %v", tpl, err)
	}
}

func TestShouldReturnNotFoundWhenNameMissing(t *testing.T) {
	_, c := newFakeServer(t, ok(`{"data":[],"pagination":{}}`))
	_, err := c.Templates.FindByName(bg, "nope")
	apiErr, isAPI := err.(*APIError)
	if !isAPI || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %v", err)
	}
}

func TestShouldPropagateListErrorOnFindByName(t *testing.T) {
	_, c := newFakeServer(t, fakeResponse{status: http.StatusBadRequest, body: `{"error":{"code":"X"}}`})
	if _, err := c.Templates.FindByName(bg, "a"); err == nil {
		t.Fatal("expected error")
	}
}

func TestShouldAddressTemplatesByID(t *testing.T) {
	fs, c := newFakeServer(t,
		ok(`{"id":"`+templateID+`","name":"boas_vindas"}`),
		ok(`{"status":"APPROVED","category":"UTILITY"}`),
		fakeResponse{status: http.StatusNoContent},
	)
	tpl, err := c.Templates.Get(bg, templateID)
	if err != nil || tpl.Name != "boas_vindas" {
		t.Fatalf("get: %+v %v", tpl, err)
	}
	status, err := c.Templates.GetStatus(bg, templateID)
	if err != nil || status.Status != "APPROVED" {
		t.Fatalf("status: %+v %v", status, err)
	}
	if err := c.Templates.Delete(bg, templateID); err != nil {
		t.Fatal(err)
	}
	expectRoute(t, fs.requests[0], http.MethodGet, "/v1/templates/"+templateID)
	expectRoute(t, fs.requests[1], http.MethodGet, "/v1/templates/"+templateID+"/status")
	expectRoute(t, fs.requests[2], http.MethodDelete, "/v1/templates/"+templateID)
}

func TestShouldCreateTemplateWithBody(t *testing.T) {
	fs, c := newFakeServer(t, fakeResponse{status: http.StatusCreated, body: `{"id":"t1","name":"n"}`})
	resp, err := c.Templates.Create(bg, &CreateTemplateRequest{Name: "n", Category: "UTILITY", Language: "pt_BR", Body: "Oi {{1}}", Samples: map[string]any{"1": "Ana"}})
	if err != nil || resp.ID != "t1" {
		t.Fatalf("unexpected %+v %v", resp, err)
	}
	req := fs.only()
	expectRoute(t, req, http.MethodPost, "/v1/templates")
	if decodeBody(t, req.Body)["body"] != "Oi {{1}}" {
		t.Fatalf("unexpected body %s", req.Body)
	}
}

func TestShouldFetchTemplateAnalytics(t *testing.T) {
	fs, c := newFakeServer(t, ok(`{"sent":10}`), ok(`{"sent":3}`))
	all, err := c.Templates.Analytics(bg, "")
	if err != nil || all["sent"] != float64(10) {
		t.Fatalf("unexpected %v %v", all, err)
	}
	one, err := c.Templates.TemplateAnalytics(bg, templateID, "7d")
	if err != nil || one["sent"] != float64(3) {
		t.Fatalf("unexpected %v %v", one, err)
	}
	expectRoute(t, fs.requests[0], http.MethodGet, "/v1/templates/analytics")
	if fs.requests[0].Query != "period=30d" {
		t.Fatalf("unexpected query %q", fs.requests[0].Query)
	}
	expectRoute(t, fs.requests[1], http.MethodGet, "/v1/templates/"+templateID+"/analytics")
	if fs.requests[1].Query != "period=7d" {
		t.Fatalf("unexpected query %q", fs.requests[1].Query)
	}
}
