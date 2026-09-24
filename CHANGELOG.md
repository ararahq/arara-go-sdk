# Changelog

## 1.0.0 (2026-09-24)

First tagged release, aligned with the AraraHQ API contract of 2026-09-24.

### Added
- `Auth.Me` (`GET /auth/me`, ADMIN key) replaces the old current-user call.
- `Messages.Get`, `Messages.SendBatch` (up to 1000 items) and `Messages.ListByBatch`.
- `Templates.Analytics`, `Templates.TemplateAnalytics` and `Templates.FindByName` (list + local filter).
- `OptOuts` resource (`/v1/opt-outs`).
- `CampaignRequest.ScheduledAt`; `interactive`, `location` and `reaction` on `SendMessageRequest`; `MessageResponse.Reason`.
- Generic `Paginated[T]` for `{data, pagination}` responses.
- Typed `*APIError` with `IsPlanFeatureLocked`, `AsPlanFeatureLocked` and `IsAuthError`.
- `Version` constant, MIT LICENSE, test suite and release workflow.

### Changed
- `Messages.Send`, `Messages.SendBatch` and `Campaigns.Create` always send an `Idempotency-Key` (UUID v4 when not given), reused across retries.
- Retries only happen for GET requests and requests carrying an `Idempotency-Key`; other writes are never repeated.
- `Templates.Get`, `GetStatus` and `Delete` take the template id (UUID), not the name.
- `Templates.List` and `SmartLinks.List` return `*Paginated[T]` and accept page parameters.
- `Error` renamed to `APIError`; 401 and 403 without an error code carry `AUTHENTICATION_ERROR`.

### Removed
- `Users` (`/users/me` is not reachable with an API key).
- `Organizations` webhook get/update (no route reachable with an API key).
- `APIKeys` (the API rejects key management with an API key).
