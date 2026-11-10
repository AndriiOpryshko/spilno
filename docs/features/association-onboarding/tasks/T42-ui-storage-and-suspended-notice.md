---
id: T42
title: "Show the storage connection, the connect flow and the «Доступ призупинено» notice"
layer: "ui"
deps: ["T41"]
acs: ["AC-09", "AC-21", "AC-25", "AC-27", "AC-28", "AC-28b"]
files_hint:
  - "apps/web/src/routes/storage.callback.tsx"
  - "apps/web/src/features/identity/storage/storage-line.tsx"
  - "apps/web/src/features/identity/storage/storage-line.test.tsx"
  - "apps/web/src/features/identity/storage/storage-line.stories.tsx"
  - "apps/web/src/features/identity/association/suspended-notice.tsx"
  - "apps/web/src/features/identity/association/suspended-notice.test.tsx"
  - "apps/web/src/features/identity/association/suspended-notice.stories.tsx"
  - "apps/web/src/routes/o.$associationId.tsx"
  - "apps/web/src/routes/o.$associationId.association.tsx"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T42 — Show the storage connection, the connect flow and the «Доступ призупинено» notice

## Why

[ADR-0008](../adr/0008-let-a-board-member-connect-the-association-storage-in-the-web-app.md), [ADR-0007](../adr/0007-enforce-a-disabled-association-as-read-only-in-the-identity-middleware.md); [PRD](../PRD.md) US-07, US-08; [sad](../sad.md) §6 Flows 10 and 11.

**Satisfies:** [PRD §5](../PRD.md) AC-09, AC-21, AC-25, AC-27, AC-28, AC-28b · **Blocked by:** [T41](./T41-ui-association-card.md)

## What

- `StorageLine` on the card page: «Сховище документів: <адреса акаунта>», or «сховище не підключене» with «Підключити сховище», which goes to `startStorageConnection`.
- `/storage/callback` — one fixed address registered with the provider — posts `completeStorageConnection` with the code and state for the association the flow started from and returns to its card; the refusal «Це особистий акаунт члена правління — підключіть акаунт ОСББ».
- In the association layout: 403 `identity.association_disabled` shows `SuspendedNotice` «Доступ призупинено: <причина>»; 409 `governance.drive_not_connected` shows «Спершу підключіть сховище документів ОСББ» with a link to the card and the note that any член правління can connect it.

## Definition of Done

- [ ] Acceptance: Component tests show the storage line in both states, the consent round trip ending in the account address, the personal-account refusal, the suspended notice with its reason and the «connect the storage first» message with its link.
- [ ] both storage states; the round trip with a mocked completion ends on the card with the address
- [ ] the personal-account refusal leaves «сховище не підключене»
- [ ] the notice shows the reason from the problem's `detail`; the not-connected message carries the link
- [ ] stories for the line and the notice pass a11y
- [ ] `mise run lint` and `mise run test` green

## Notes

The upload control that triggers `governance.drive_not_connected` is board-protocol's ui task T36; the message handling lives in the layout so it works for it unchanged. Reuses `InlineAlert`, `Button`, `Card`.
