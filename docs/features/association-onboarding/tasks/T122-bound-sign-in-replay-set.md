---
id: T122
title: "Bound the set of used sign-in states"
layer: "app"
deps: []
acs: ["AC-05"]
files_hint:
  - "services/core/internal/identity/app/signin.go"
  - "services/core/internal/identity/state_bounded_test.go"
  - "services/core/internal/identity/signin_test.go"
owner: "Andrii Opryshko"
estimate: "S"
status: "todo"
---

# T122 — Bound the set of used sign-in states

## Why

[Second re-review of 2026-11-17](../_review/review-2026-11-17-2.md), S01. `markUsed` runs before the code exchange, so a callback with a valid sealed cookie and a bogus code adds an entry for ten minutes; anyone gets a fresh cookie from the public start. The set has no cap and every callback scans it under one mutex.

**Satisfies:** [PRD §5](../PRD.md) AC-05 · **Blocked by:** —

## What

- The set is capped with the eviction the consent states already use; the sweep of expired entries runs only past the cap.
- The comment says what is kept.

## Definition of Done

- [ ] Acceptance: A test sends more start-and-callback pairs with a bogus code than the cap and sees the set stay at the cap, a sign-in started before the flood still complete, and a replay of a completed state still refused.
- [ ] `state_bounded_test.go`: the flood, the size, the earlier sign-in, the replay
- [ ] `mise run lint` and `mise run test` green

## Notes

Losing an old entry reopens only the replay of a code the provider has already consumed.
