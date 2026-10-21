---
status: accepted
date: 2026-10-21
---

# 0002 Bounded contexts inside one Go module

## Context and problem statement

The core service is a modular monolith. Contexts must stay independent so that a context can
later move into its own service, and so that generated code never leaks across contexts.

## Decision outcome

Each context is `internal/<bc>/` with `domain/`, `app/`, `ports/`, `adapters/http{,/gen}/`,
`adapters/postgres{,/queries,/sqlcgen}/` and a root `module.go`. Only the root package may be
imported from outside the context.

Enforcement:

1. depguard (golangci-lint): `domain-pure` allows only the standard library and uuid in
   `*/domain/**`; one `<bc>-private` rule per context denies `internal/<bc>/` sub-packages
   everywhere else.
2. The petforge plugin hook blocks such imports while Claude edits files, before lint runs.

Go's `internal/` does not help here: siblings under `internal/` may import each other freely.

### Consequences

- Good: violations fail CI and are caught at edit time.
- Bad: each new context needs its depguard rule; the plugin generates it.
