---
status: accepted
date: 2026-10-21
---

# 0003 Spec-first API with per-context code generation

## Context and problem statement

Go server and TS client must agree on the contract, and adding an endpoint should be a
mechanical, repeatable sequence that a plugin can drive.

## Considered options

- Spec-first: OpenAPI file, oapi-codegen for Go, orval for TS.
- Code-first: huma derives the spec from Go types.
- No codegen: hand-written handlers and client.

## Decision outcome

Spec-first. One `services/core/api/openapi.yaml`; every operation carries one context tag.
Each context generates its own server with `include-tags: [<bc>]`, which also prunes unused
schemas. orval generates TanStack Query hooks per tag into `apps/web/src/api/gen`.

Wiring rules, verified in a spike:

- Each context registers in its own `r.Group` with a validator built from its own embedded spec;
  `r.Mount` returns 404 because generated handlers use absolute paths.
- The request validator, strict request/response error handlers and the param-binding error
  handler all go through `internal/platform/httpapi`, so every error is problem+json
  (422 for validation, 404/405 for routing). Defaults would answer `400 text/plain`.
- `format: uuid` is registered with kin-openapi explicitly; it is not validated otherwise.
- Enum constants are always prefixed with their type name.
- orval uses `forceSuccessResponse` so non-2xx responses throw with a typed `info` body.

### Consequences

- Good: the build breaks until every new operation has a handler, which works as a checklist.
- Bad: `Problem` is generated once per context; the shared writer in `httpapi` mirrors it.
- Bad: OpenAPI 3.1 parameter validation messages from kin-openapi are less readable than 3.0.
- Bad: no response validation (oapi-codegen issue #1038).
