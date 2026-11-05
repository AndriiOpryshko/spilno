---
status: accepted
date: 2026-11-05
feature: association-onboarding
sad_section: "§8"
deciders: ["Andrii Opryshko", "Claude (ao-architecture)"]
---

# 0008 Let a board member connect the association storage in the web app

## Context and problem statement

Board-protocol [ADR-0005](../../board-protocol/adr/0005-store-files-in-association-owned-google-drive-folder.md) authorizes the association's Google Drive once, through a developer command (task T12). This feature lets the board connect it itself, with the association's own account, not a member's personal one (PRD US-08, AC-27, AC-28b).

## Decision drivers

- PRD AC-27, AC-28, AC-28b; brief §9 «Drive on the association's own account».
- Board-protocol ADR-0005: `drive.file` scope, AES-GCM refresh token in `google_connections`.

## Blast radius

- Irreversible: no.
- Multi-context: yes — identity owns the connection, governance uses it and lists members' sign-in addresses.
- Legitimate alternative: yes — keep the developer command.
- Touches the database: yes — `google_connections` rows written from the web flow.

## Considered options

1. **An in-app OAuth consent started by a current член правління from the association's card; the callback stores the encrypted refresh token in `google_connections` and creates the `spilno/` folder; refused with `identity.storage_account_personal` when the Google address is a member's sign-in address.**
2. **Keep the T12 developer command.**

## Decision outcome

**Chosen:** Option 1. No developer step per association; the personal-account rule is checked where the account is known.

### Consequences

- Good: replaces T12; reconnecting replaces the row as before.
- Bad: one more OAuth redirect URI to register; an unknown personal address is not detectable (SAD §11).
- Neutral: scope, encryption and folder layout of board-protocol ADR-0005 are unchanged.

### Migration and rollback

- Migration: none for the table; `google_connections` gains the row-level-security policy of [ADR-0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md).
- Rollback: fall back to the developer command; rows stay.

## Links

- PRD: [PRD.md](../PRD.md) — US-08, AC-27, AC-28, AC-28b
- SAD: [sad.md](../sad.md) §8
- Related ADR: board-protocol [ADR-0005](../../board-protocol/adr/0005-store-files-in-association-owned-google-drive-folder.md); [0003](./0003-enforce-association-isolation-with-postgres-row-level-security.md)
