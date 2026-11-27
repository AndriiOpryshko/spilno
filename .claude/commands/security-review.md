---
description: Security pass over the branch diff, tuned to spilno's threat model (overrides the built-in /security-review).
argument-hint: "[base, default origin/main]"
---

# /security-review (spilno)

A project copy of the built-in `/security-review`. The generic pass (injection, secrets, path traversal,
SSRF, unsafe deserialization) still runs; the rules below come **first**, because they are where this repo
has actually leaked or nearly leaked.

## Scope

1. `git fetch -q origin` then `git diff ${ARGUMENTS:-origin/main}...HEAD` (fall back to `main` if offline).
2. Read every changed file in full where a hunk touches a rule below; do not judge a guard from the hunk alone.
3. Ignore generated code (`*.gen.go`, `sqlcgen/`, `apps/web/src/api/gen/`, `routeTree.gen.ts`) except to
   check that what generated it (`openapi.yaml`, `queries/*.sql`) is safe.

## Threat model (spilno-specific, check each one explicitly)

**S1. Cross-association leak (P0).** One ОСББ must never read or write another's data.
- Every read or write of an association-scoped table runs inside `platform/db.InAssociation`
  (`services/core/internal/platform/db/db.go`), which sets the association the RLS policies read.
  A query on the bare pool, a `SET` done by hand, or a new code path that skips `InAssociation` is P0.
- A new association-scoped table needs `ENABLE ROW LEVEL SECURITY` + a policy in the same migration as the
  table (see `migrations/00045_enable_rls_governance.sql`) and a grant to `spilno_app` only.
  Anything that would need `BYPASSRLS`, `SECURITY DEFINER`, or the owner role at runtime is P0.
- The association id comes from the path (`/<context>/associations/{association_id}/...`), never from the body.
- New scoped data must have a `testdb.NewTwoAssociations(t)` test proving association B cannot see A's rows.

**S2. Authorization order and existence-hiding (P0/P1).**
- A scoped operation checks membership (`identity.not_current_member`, then `identity.association_disabled`)
  **before** any resource is read; an existing and a missing id return the same body. A 404 that reveals
  existence to a non-member, or a different error body for them, is P1.
- No new `security: []` in `services/core/api/openapi.yaml` beyond health and the sign-in redirect/callback.
- The dev sign-in shim (`internal/identity/adapters/oidc/dev.go`) must stay behind `SPILNO_DEV_SIGN_IN=true`;
  anything that enables it by default or widens it is P0.

**S3. Secrets, internal addresses and PII on the wire or in logs (P1).**
- RFC 9457 `detail`/`errors[]` and log fields never carry: upstream URLs or hosts (e.g. the NBU / MCP
  address), DSNs, tokens, Google Drive file or revision ids, emails, names, IBANs, money amounts tied to a
  person. Wrap upstream errors; log a code, not the raw error string, when the raw one can contain a URL.
- `example:` blocks in the contract use `@example.test` and `Test User` only.

**S4. Outbound calls (P1).** `NBU_RATES_MCP_URL` and other outbound targets come from config, never from a
request. Every outbound call has a timeout and a bounded response size.

**S5. Files (P1).** Uploads are size- and type-checked; downloads go out only as
`Content-Disposition: attachment` through the API, with a sanitized filename.

## Output

For every finding: `file:line`, rule id (S1–S5 or a generic class), severity (P0 blocks merge, P1 fix in
this PR, P2 note), the concrete exploit path in one sentence, and the fix. Then a one-line verdict:
`SECURITY: BLOCK` (any P0), `SECURITY: FIX` (P1 only) or `SECURITY: PASS`.

Write the result to `docs/features/<slug>/_review/security-<YYYY-MM-DD>.md` when the branch belongs to a
feature, so a fresh session can pick the fixes up (persist-then-pick-up). Do not fix anything in this pass.

## Honest limit

An LLM pass catches patterns, not everything. It complements, and does not replace, `golangci-lint` (gosec),
the RLS integration tests and secret scanning in CI.
