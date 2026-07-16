# arara-go-sdk

Official Go SDK for the [AraraHQ](https://ararahq.com) API — the best WhatsApp API for Brazilian developers.

Zero external dependencies. Standard library only (`net/http`, `encoding/json`, `context`). Requires Go 1.22+.

## Install

```bash
go get github.com/ararahq/arara-go-sdk
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

	fmt.Printf("message %s status %s\n", resp.ID, resp.Status)
}
```

## Configuration

```go
client, err := arara.NewClient(
	"ara_live_your_key_here",
	arara.WithBaseURL("https://api.ararahq.com"),
	arara.WithTimeout(15*time.Second),
	arara.WithMaxRetries(5),
)
```

`429` and `5xx` responses (and network failures) are retried with exponential backoff, honoring the `Retry-After` header.

## Resources

`client.Messages`, `client.Templates`, `client.Users`, `client.Organizations`, `client.APIKeys`, `client.Contacts`, `client.Conversations`, `client.Wallet`, `client.Numbers`, `client.SmartLinks`, `client.Campaigns`.

Every method takes a `context.Context` as its first argument.

### Idempotency

`Messages.Send` accepts an optional `Idempotency-Key` header:

```go
client.Messages.Send(ctx, req, arara.SendOptions{IdempotencyKey: "order-42"})
```

`Campaigns.Create` requires one — a UUID v4 is generated automatically when you don't pass `CampaignCreateOptions{IdempotencyKey: ...}`.

## Key scope

API key scope is enforced server-side: READ keys do GET only; SEND keys add POST on `/messages` and `/campaigns`; everything else requires an ADMIN key.

## Errors

Non-2xx responses return a typed `*arara.Error` with `StatusCode`, `Code`, `Message`, `Details`, and `RetryAfter`.
