---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0017 Store protocol text as fragments with slot placeholders, edited in spilno with a revision number

## Context and problem statement

The product owner decided on 2026-10-25 that the wording is edited in spilno (PRD §8, closed OQ), by hand or in the conversation with the помічник протоколу (US-04). A second save of an old version is refused (AC-11b), who changed the text is recorded (AC-11), every slot keeps its value and source through a text edit (AC-11), a poll round freezes the current text (AC-13, AC-17) and the print copy freezes it again (AC-25, AC-26). ADR 0006 put the wording in a shared Google Doc, which could not honour AC-11b and needed Doc sharing for members; this ADR supersedes ADR 0006. The idea of text as fragments with placeholders is ported from the other SAD's ADR 0011 (`adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md`).

## Decision drivers

- AC-11, AC-11b: hand edits keep slots; a stale save is refused with `governance.draft_changed`.
- AC-06, AC-43: a number the помічник proposed reaches print only after a person confirms it; a new document version makes it unconfirmed again.
- AC-13, AC-17, AC-25, AC-26: poll rounds and print copies reference an exact, frozen text.
- PRD §6.1 abuse case "Instructions hidden in a document": generated text must not smuggle facts past the confirmation gate.

## Blast radius

- Irreversible: yes — the stored shape of the legal text, the editor UI and the renderer all follow it; changing it later rewrites every revision.
- Multi-context: yes — `apps/web` (editor with slot chips) and `governance` (core and worker); the shape is part of the OpenAPI contract.
- Legitimate alternative: yes — free text with numbers inlined and checked afterwards, or the Google Docs editing of ADR 0006.
- Touches the database: yes — `protocol_revisions` (append-only), `protocols.revision`.

## Considered options

1. **Fragments with placeholders, edited in spilno** — each agenda item's text is a list of plain fragments and references to slots or meeting facts; values are substituted when shown or rendered.
2. **Free text, checked afterwards** — the model and the member write ordinary text with values inlined; a pattern check compares numbers with slots (that check is AC-12, Stage 2).
3. **Google Docs while in чернетка** (ADR 0006) — rejected after the product owner's 2026-10-25 decision.

## Decision outcome

**Chosen:** option 1. A slot change updates every place the value appears, so text and numbers cannot diverge, and no Doc sharing or Drive revision tracking is needed.

- Shape per agenda item: `[{text: "…"} | {slot: <slot id>} | {meeting: date | chair | members}]`, stored as jsonb. The помічник is asked for this structure (ADR 0005). A deterministic validator in `governance/domain` turns any fact in a generated plain fragment (digits, Ukrainian number words, dates, currency marks, document numbers, roster names) into a proposed new slot with origin `ai`, unconfirmed — so it falls under the one Stage 1 print gate (AC-06). For hand-typed text the validator only highlights such fragments; making it blocking is AC-12 (Stage 2).
- Every change sends the protocol's `revision`; `app/` compares and increments it in one `UPDATE … WHERE revision = $expected`; a mismatch answers `governance.draft_changed` with the newer revision (AC-11b). Each accepted change appends a `protocol_revisions` row (revision, agenda and text jsonb, slot values snapshot, changed by, changed at) — this is "who changed it" (AC-11).
- A poll round (AC-13) and a print copy (AC-25, ADR 0018) each reference one revision. Any change to text, agenda or a value while готовий до голосування or готовий до підпису returns the protocol to чернетка in the same transaction (AC-17, AC-26, AC-43).
- The web editor is a structured text field with slot chips (shadcn primitives); chips show origin and confirmation (US-03).

### Consequences

- Good: AC-11b holds for the whole draft — no decision override; prompt-injected numbers become unconfirmed slots.
- Good: one renderer formats values (amounts in digits and words, dates) for screen and print copy alike.
- Bad: a structured editor in `apps/web` instead of a plain textarea; the validator's Ukrainian word lists need their own tests, and a false positive forces rephrasing.
- Neutral: plain prose without facts stays free to write.

### Migration and rollback

- Migration: `protocols.revision int NOT NULL DEFAULT 1`; new table `protocol_revisions` (protocol id, revision, body jsonb, changed by, changed at, unique (protocol id, revision)); no existing rows.
- Rollback: revisions are kept; a later free-text editor renders the last revision to plain text.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-03, US-04; AC-06, AC-11, AC-11b, AC-13, AC-17, AC-25, AC-26, AC-43; §6.1 abuse cases; §8 closed OQ "Where is the wording edited"
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §6, §8
- Related ADR: supersedes [0006 Edit wording in Google Docs](./0006-edit-protocol-wording-in-google-docs-and-freeze-it-into-postgres-on-transition.md) · [0018 Render the print copy](./0018-render-the-print-copy-from-a-pinned-google-docs-template-in-a-worker-job.md) · idea from [other SAD ADR 0011](../adr/0011-store-decision-text-as-fragments-with-slot-placeholders.md)
