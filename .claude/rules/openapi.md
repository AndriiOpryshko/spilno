# API contract rules — baseline

<!-- Bootstrapped by sdlc/plugin/skills/api-forge on the first HTTP run (board-protocol, 2026-10-27). Edit freely. -->
<!-- These rules are an opinionated default — see "Defaults" in SKILL.md. A team edit here, or an ADR, is the only way to deviate; api-forge reads this file on every run and records deviations in api-sync-report.md. -->

## Where the contract lives

- `docs/features/<slug>/contracts/openapi.yaml` is the source of truth for the HTTP surface. Handlers, clients, mocks and contract tests follow it — never the other way round.
- The contract is **derived** (`data-model.md` + `sad.md` §6 sequences + `PRD.md` AC) by `/sdlc:api-forge <slug>`. After a hand edit or an upstream change, run `/sdlc:api-forge <slug> --reconcile` so `api-sync-report.md` stays true.
- `info.version` bumps only with an explicit CHANGELOG line, never silently.

## Hard rules

- OpenAPI **3.1.0**. Nullability via `type: [string, null]`; `nullable: true` (3.0 style) is forbidden.
- Every error response is the envelope `{code, message, details?}`. `code` is `module.error_name` in snake_case (`lesson.not_found`, `course.not_methodist`) — never free text like `{"error": "failed"}`.
- One error `code` in the contract = one definition in the repo's error registry (constants / enum / sentinel module — whatever the repo uses), returning the **same string** the contract advertises.
- Shared types via `$ref` only; no inline duplicated schemas.
- No `additionalProperties: true` on response shapes. Exception only by ADR (e.g. a polymorphic payload) and noted in the report.
- No real PII in `example` blocks: `<name>@example.test`, `+380 00 000 00 00`, `Test User`.

## Defaults

- URL versioning `/api/v1/...`; never `?v=2` query versioning.
- `BearerAuth` global; a public endpoint declares explicit `security: []`.
- Lists paginate by **cursor** (`?after=&before=&limit=`), wrapped in `{items, has_next, has_prev, next_cursor}`; never offset.
- JSON keys snake_case. Schema names use the domain language of `data-model.md` (`Lesson`, `CreateLessonRequest`) — not a Go/TS/Python idiom, not a driver's error type.
- A mutating endpoint behind a retry or an async actor requires an `Idempotency-Key` header (state the TTL).
- Status by class: `400` validation, `401` no/invalid token, `403` not allowed, `404` existence-hiding, `409` state or uniqueness conflict, `429` rate limit, `5xx` server. Every `4xx`/`5xx` carries the envelope.

## Checks

- `spectral lint docs/features/<slug>/contracts/openapi.yaml` is in the project's check target (`make check` or stack-equivalent).
- A change to the contract and the matching handler/client land in the same PR, with `api-sync-report.md` re-run (`--reconcile`) and its `**Verdict:**` at `PASS` or `PASS (with deferrals)`.

## spilno specifics

These override the baseline above; each names the decision behind it.

- **One contract per repo.** The live contract is `services/core/api/openapi.yaml` (repo ADR 0003). A feature's
  `docs/features/<slug>/contracts/openapi.yaml` is a self-contained **delta** (new operations, new schemas,
  changed schemas) that `implement-tasks` merges into it; then `mise run gen`.
- **Paths are `/<context>/...`, no `/api/v1` prefix.** Every operation has exactly one tag, its bounded context
  (CLAUDE.md, repo ADR 0003). `/api` is the SPA proxy prefix stripped before the core API (`apps/web/vite.config.ts`,
  orval `baseUrl: '/api'`); a breaking change gets a new operation, not a path version.
- **Errors are RFC 9457 `application/problem+json`** (`internal/platform/httpapi`, CLAUDE.md): `{type, title, status,
  detail?, instance?, errors?[], code?}`. The envelope maps as `code` → `code` (extension member, `<context>.<error_name>`),
  `message` → `detail`, `details` → `errors[]` (`{pointer, detail}`). Validation is `422` and names every broken field
  at once; `400` is only for a body that cannot be decoded. A `422` produced by the shared request validator carries
  no `code`; a `422`/`409` produced by a context carries its `code`.
- **Auth is a session cookie**, not a bearer token (board-protocol ADR-0003): global `SessionCookie`; a public
  operation (`/system/health`, the sign-in redirect and callback) declares `security: []`. No session → `401
  identity.session_required`; a session whose person is not a current член правління → `403
  identity.not_current_member` before any resource is read, the same body for an existing and a missing id.
- **Pagination** keeps the existing shape: `?cursor=&limit=` (limit 1–100, default 20), response `{items, next_cursor}`
  with `next_cursor: null` on the last page; opaque cursor over `(created_at, id)`, newest first (board-protocol SAD §8).
- **Optimistic concurrency instead of `Idempotency-Key`** on protocol mutations: each carries `protocol_revision`;
  a stale or replayed write answers `409 governance.draft_revision_conflict` (board-protocol ADR-0009). An
  `Idempotency-Key` is required only where a mutation has an async follow-up and no revision guard.
- **Money** is integer minor units plus `Currency` (`amount_cents` in finance, `amount_kopiyky` in governance).
- **Files** are uploaded as `multipart/form-data` and downloaded only through the API as
  `Content-Disposition: attachment`; storage ids (Google Drive file and revision ids) never appear on the wire.
- **Lint:** no spectral target in `mise` yet; until one is wired, the delta is checked with
  `npx @stoplight/spectral-cli lint` and by `mise run gen` after the merge.
