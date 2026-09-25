# arara-go-sdk

Official Go SDK for the [AraraHQ](https://ararahq.com) API — the best WhatsApp API for Brazilian developers.

Zero external dependencies. Standard library only (`net/http`, `encoding/json`, `context`). Requires Go 1.22+.

## Install

```bash
go get github.com/ararahq/arara-go-sdk@latest
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	arara "github.com/ararahq/arara-go-sdk"
)

func main() {
	client, err := arara.NewClient("ara_live_your_key_here")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Messages.Send(context.Background(), &arara.SendMessageRequest{
		Receiver:          "+5511999998888",
		TemplateName:      "boas_vindas",
		TemplateVariables: []string{"Micael"},
	})
	if err != nil {
		log.Fatal(err)
	}

	if resp.ID != nil {
		fmt.Printf("message %s status %s\n", *resp.ID, resp.Status)
	}
}
```

`MessageResponse.ID`, `MessageResponse.Reason` and `TemplateStatus.Category` are `*string` because the API may return `null`; check for nil before dereferencing.

`Receiver` accepts `whatsapp:+5511...`, `+5511...` or digits only; the API validates it.

## Configuration

```go
client, err := arara.NewClient(
	"ara_live_your_key_here",
	arara.WithBaseURL("https://api.ararahq.com"),
	arara.WithTimeout(15*time.Second),
	arara.WithMaxRetries(5),
)
```

`429`, `5xx` and network failures are retried with exponential backoff, honoring `Retry-After`. Only GET requests and requests carrying an `Idempotency-Key` are retried; any other write is never repeated.

## Resources

| Field | Endpoints | Key |
|---|---|---|
| `Messages` | `Send`, `SendBatch` (≤1000), `Get`, `ListByBatch` | SEND / READ |
| `Templates` | `List` (paginated), `Create`, `Get`/`GetStatus`/`Delete` **by id (UUID)**, `FindByName`, `Analytics`, `TemplateAnalytics` | READ / TEMPLATES_WRITE |
| `Campaigns` | `Create`, `List`, `Get`, `Estimate`, `Cancel` | READ / CAMPAIGNS_SEND |
| `Numbers` | `/v1/organizations/me/numbers` | READ (writes: ADMIN) |
| `SmartLinks` | `List` (paginated), `Create`, `Update`, `Stats` | ADMIN |
| `Contacts` | list, batch import, stats, tags, get/update | **ADMIN** for reads |
| `Conversations` | list, messages, reply, status, window status | **ADMIN** |
| `Wallet` | transactions, auto-recharge | **ADMIN** |
| `OptOuts` | list, get, create, delete | **ADMIN** |
| `Auth` | `Me` (`GET /auth/me`) | **ADMIN** |

Every method takes a `context.Context` as its first argument.

`Templates.FindByName` calls `List` with the `name` filter and matches the first page locally; use the returned `ID` for the id-based calls.

### Pagination

`Templates.List` and `SmartLinks.List` return `*arara.Paginated[T]`:

```go
page, err := client.Templates.List(ctx, arara.TemplateListParams{Status: "APPROVED", Page: 0, Size: 50})
for _, tpl := range page.Data {
	fmt.Println(tpl.ID, tpl.Name)
}
fmt.Println(page.Pagination.TotalPages)
```

Campaigns and wallet transactions use the API's `{content, totalPages, totalElements}` shape; contacts use `{contacts, total, page, size, totalPages}`.

### Idempotency

`Messages.Send`, `Messages.SendBatch` and `Campaigns.Create` always send an `Idempotency-Key`. When you don't pass one, a UUID v4 is generated per call and reused on every retry of that call.

```go
client.Messages.Send(ctx, req, arara.SendOptions{IdempotencyKey: "order-42"})
client.Campaigns.Create(ctx, campaign, arara.CampaignCreateOptions{IdempotencyKey: "black-friday"})
```

## Key permissions

Enforced server-side. READ keys can only GET messages, campaigns, templates, numbers, automations, flows and charges. Contacts, conversations, wallet, opt-outs, smart links and `Auth.Me` require an ADMIN key. API key management and the organization webhook are not available through an API key.

## Errors

Non-2xx responses return `*arara.APIError` with `StatusCode`, `Code`, `Message`, `Details` and `RetryAfter`:

```go
_, err := client.Templates.Get(ctx, id)
var apiErr *arara.APIError
if errors.As(err, &apiErr) {
	log.Printf("%s: %s", apiErr.Code, apiErr.Message)
}
if lock, ok := arara.AsPlanFeatureLocked(err); ok {
	log.Printf("feature %s needs plan %s (current %s)", lock.Feature, lock.UpgradeTo, lock.CurrentPlan)
}
if errors.Is(err, context.Canceled) {
	log.Print("canceled by the caller")
}
```

- `arara.IsPlanFeatureLocked(err)`: `403 PLAN_FEATURE_LOCKED`.
- `arara.IsAuthError(err)`: 401, the key was rejected.
- `arara.IsForbidden(err)`: 403 without an error code (`Code == "FORBIDDEN"`). Ambiguous by design of the API: a key without permission for the route, or a resource of another organization (`GET /v1/messages/{id}` answers an empty 403).
- `arara.IsNotFound(err)`: 404 (`Code == "NOT_FOUND"` when the body is empty).

## License

MIT
