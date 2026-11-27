---
id: T11
title: "Prove the chain end to end against the real nbu-rates binary and a fake NBU"
layer: "tests"
deps: ["T10"]
acs: ["AC-11", "AC-01", "AC-10"]
files_hint:
  - "services/core/internal/finance/rates_e2e_test.go"
owner: "Andrii Opryshko"
estimate: "M"
status: "done"
---

# T11 — Prove the chain end to end against the real nbu-rates binary and a fake NBU

## Why

[SAD](../sad.md) §11: the adapter's stub server can drift from the real tool contract; this test pins the two together.

**Satisfies:** [PRD §5](../PRD.md) AC-11, AC-01, AC-10 · **Blocked by:** T10

## What

- `go build -C <repo>/mcp/nbu-rates -o <tmp>/nbu-rates .` inside the test; skip with a clear message only if the directory is absent.
- Start the binary with `-http 127.0.0.1:<free port>`; wait on `/healthz`; kill it in `t.Cleanup`.
- `finance.New(pool, finance.Config{RatesMCPURL: …})` behind the usual guarded test server.

## Definition of Done

- [ ] Acceptance: One test builds `mcp/nbu-rates`, starts it on a free loopback port with `NBU_BASE_URL` pointing at an `httptest` NBU, creates a EUR contract through the finance module's HTTP surface and reads `uah_equivalent` = 6316388 at `50.5311`; a second contract for a day the fake does not publish is created with `uah_equivalent: null` and the later call answers `409 finance.rate_not_published`. No request leaves the machine.
- [ ] end to end: the two scenarios of the DoD
- [ ] `mise run lint` and `mise run test` green
