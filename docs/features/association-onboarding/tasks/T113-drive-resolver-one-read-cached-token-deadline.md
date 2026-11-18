---
id: T113
title: "Resolve an association's Drive storage with one read, a reused token and a deadline"
layer: "app"
deps: ["T112"]
acs: ["AC-26", "AC-28"]
files_hint:
  - "services/core/cmd/api/main.go"
  - "services/core/internal/identity/google.go"
  - "services/core/internal/identity/app/google_connections.go"
  - "services/core/internal/governance/app/files.go"
  - "services/core/internal/governance/app/cases.go"
  - "services/core/internal/governance/drive_wiring_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T113 — Resolve an association's Drive storage with one read, a reused token and a deadline

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S03. Every upload and download asks Google's token endpoint for a new access token and reads `google_connections` two or three times; the Drive client has no deadline; an unreadable token (a rotated key) answers 500 instead of the advertised 503.

**Satisfies:** [PRD §5](../PRD.md) AC-26, AC-28 · **Blocked by:** [T112](./T112-wire-consent-drive-address-and-test-api-wiring.md)

## What

- Identity returns the token source and the folder id of a connection from one read.
- The token source is reused per association and dropped when the connection row changes (keyed by the connection's id).
- The Drive client carries a deadline; a resolver failure that is not «no connection» answers 503 `governance.storage_unavailable`.

## Definition of Done

- [ ] Acceptance: Two downloads of one association make one call to the token endpoint, a Drive that never answers ends the request with 503 within the configured deadline, and an unreadable token answers 503 `governance.storage_unavailable`.
- [ ] module test: two requests, one token call; a reconnect makes the next request use the new connection
- [ ] module test: a hanging fake Drive answers 503 within the deadline
- [ ] module test: `ErrGoogleTokenUnreadable` answers 503 on upload and on download
- [ ] `mise run lint` and `mise run test` green

## Notes

Lands after [T112](./T112-wire-consent-drive-address-and-test-api-wiring.md), which moves the resolver into the constructor this task changes.
