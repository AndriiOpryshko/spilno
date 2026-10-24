---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0006 Edit protocol wording in Google Docs and freeze it into Postgres on every transition

## Context and problem statement

Members edit the protocol wording by hand or in the conversation (US-04). PRD §8 left open where; the author proposed Google Docs. The PRD needs a frozen numbered version for the попереднє голосування (AC-13, AC-17), a text frozen for printing (AC-23, AC-26), numbers in the text equal to the slots (AC-12), who changed the text (AC-11), and AC-11b: a second save of an old version is refused. Docs edits are live and merged, so AC-11b cannot hold for the wording as written.

## Decision drivers

- AC-12, AC-13, AC-17, AC-26: the frozen version is what members vote on and what is printed.
- QG-1: the legal text is the frozen version, not a live document.
- QG-3: only current members of the board may edit (AC-32).
- PRD §8 OQ: the author prefers a shared online editor.

## Blast radius

- Irreversible: yes — the editing surface shapes the web UI, the worker jobs and how versions are stored.
- Multi-context: yes — web, `core`, `worker`, and `identity` (members' Google accounts).
- Legitimate alternative: yes — an editor in spilno with versioned blocks and slot tokens.
- Touches the database: yes — frozen versions table, Doc id and revision per version.

## Considered options

1. **Google Docs while in чернетка, frozen into Postgres on every transition** — the Doc is locked in frozen states.
2. **Editor in spilno** — blocks with slot tokens, every save a version with an expected-version check.

## Decision outcome

**Chosen:** option 1. The worker creates one Doc per draft in the protocol folder (ADR 0004), writes the assistant's wording, and marks each slot value with a Docs named range. While the draft is чернетка the Doc is shared as editor, without notification e-mails, with the Google accounts of the members whose term is in force; the worker revokes access when a term ends. On готовий до голосування or готовий до підпису, `core` exports the Doc text at its current revision (Docs API), compares every named range and every other number in the text with the slots (`governance.text_slot_mismatch`), stores the text as `protocol_versions` row N with the Doc revision id and sha256, and enqueues `apply_drive_restrictions` to set the Doc read-only. Editing in spilno ("Редагувати текст") returns the draft to чернетка and the same job lifts the restriction (AC-17, AC-26). If Docs is unavailable at export, the transition is refused with `governance.storage_unavailable` (SAD §6 Flow 8). Who changed the text comes from the Drive revision history and is copied into the frozen version.

**Decision override, PRD AC-11b:** for wording, "refuse the second save of an old version" becomes "real-time co-editing without losing either change". AC-11b holds as written for slots, agenda and meeting facts edited in spilno, which carry an optimistic version number. The PRD is to be updated (SAD §11).

### Consequences

- Good: a familiar editor with co-editing; less UI code in `apps/web`.
- Good: the legal text and its hash live in Postgres, independent of later Doc changes.
- Bad: scopes `drive.file` + Docs; sharing and revoking are a worker job, so revocation lags by up to one worker cycle.
- Bad: a member can delete a named range; the check then reports the slot as missing from the text and blocks.
- Neutral: the frozen version is plain text with structure; the PDF is rendered from it, not from the Doc.

### Migration and rollback

- Migration: new table `protocol_versions` (number, text, doc revision id, sha256, frozen by, frozen at); `protocols.doc_id`.
- Rollback: moving to an in-app editor later imports the last frozen version as block text; frozen versions are kept as they are.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-04, US-05, AC-11, AC-11b, AC-12, AC-13, AC-17, AC-26; §8 hand-editing OQ
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §11
- Related ADR: [0004 Files on Google Drive](./0004-store-case-files-and-signed-scans-on-the-associations-google-drive-with-hashes-in-postgres.md) · [0003 Sign in with Google](./0003-sign-in-with-google-in-a-new-identity-context.md)
