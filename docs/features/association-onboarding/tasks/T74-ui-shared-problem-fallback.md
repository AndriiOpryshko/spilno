---
id: T74
title: "Give every page a loading state and one fallback for a failed request"
layer: "ui"
deps: ["T68"]
acs: ["AC-05", "AC-07"]
files_hint:
  - "apps/web/src/routes/index.tsx"
  - "apps/web/src/routes/o.$associationId.tsx"
  - "apps/web/src/routes/o.$associationId.association.tsx"
  - "apps/web/src/routes/applications.index.tsx"
  - "apps/web/src/routes/storage.callback.tsx"
  - "apps/web/src/features/identity/applications/application-detail.tsx"
  - "apps/web/src/features/identity/applications/application-detail.test.tsx"
  - "apps/web/src/features/identity/storage/storage-line.tsx"
  - "apps/web/src/features/identity/storage/storage-line.test.tsx"
  - "apps/web/src/features/identity/shell/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T74 — Give every page a loading state and one fallback for a failed request

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), W2. [PRD](../PRD.md) AC-07: a refusal looks the same everywhere. Several pages stay on «Завантаження…» forever, show an empty list while loading, send a member to `/applications/` on a 5xx, send any failure to sign-in, or render blank.

**Satisfies:** [PRD §5](../PRD.md) AC-05, AC-07 · **Blocked by:** [T68](./T68-ui-association-shell-card-notice.md)

## What

- One shared fallback used by `routes/index.tsx`, `o.$associationId.tsx`, `o.$associationId.association.tsx`, `applications.index.tsx`, `application-detail.tsx`, `storage.callback.tsx` and `storage-line.tsx`.
- `routes/index.tsx` redirects with `<Navigate>`, not `window.location.replace` during render.
- `storage-line.tsx` shows a pending state and a message for refusals other than the personal account.

## Definition of Done

- [ ] Acceptance: Component tests show each listed page rendering a skeleton while loading, `/sign-in` on 401, `/no-access` on the access codes and an alert with «Спробувати ще раз» on 5xx or a network error.
- [ ] a test per page for loading, 401, 403 and 5xx
- [ ] the storage callback without its sessionStorage key explains what happened and links back to the card
- [ ] `mise run lint` and `mise run test` green

## Notes

Reuses `InlineAlert`, `Skeleton`, `Button`.
