---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0018 Render the print copy from a pinned Google Docs template in a worker job

## Context and problem statement

Moving to готовий до підпису renders a PDF print copy from the current text, with the protocol number, the meeting form and signers from the settings and blank fields for members present and the vote count per item (AC-25, AC-45); a later edit voids it and keeps it in history (AC-26); a withdrawn protocol keeps it as never signed (AC-34c). The rendering technology was left open in this SAD (job `render_pdf` had no engine), and ADR 0006, now superseded by ADR 0017, no longer gives Google Docs a role in editing. The idea of rendering through a pinned Docs template is ported from the other SAD's ADR 0013 (`adr/0013-render-signing-copies-from-google-docs-templates.md`).

## Decision drivers

- PRD §6: print copy (PDF) rendering ≤ 5 s p95.
- AC-25, AC-26, AC-34c: the system knows exactly which copy was issued, for which revision, and whether it is void.
- QG-1: the issued file is verifiable later (sha256).
- One developer; Cyrillic typography; files already live in the association's Drive (ADR 0004).

## Blast radius

- Irreversible: no — the renderer sits behind `governance/ports.Renderer`; another engine is a new adapter.
- Multi-context: yes — `governance` (core enqueues, worker renders) and `identity` (the association's Google token); the web shows the job state and downloads the PDF.
- Legitimate alternative: yes — Typst in the worker, or HTML → PDF through a headless-Chromium container (Gotenberg).
- Touches the database: yes — `print_copies` rows (revision, number, Drive file id, sha256, state issued / void / never signed).

## Considered options

1. **Google Docs template + Drive PDF export** — the worker copies a pinned template Doc, fills it through the Docs API, exports PDF, deletes the intermediate Doc.
2. **Typst in the worker** — a template in the repo rendered by the `typst` binary.
3. **HTML → PDF through Gotenberg** — a headless-Chromium container in compose.

## Decision outcome

**Chosen:** option 1. No new binary or container, Cyrillic and fonts work as in any Doc, and the board can adjust the layout in an editor it knows; Google is already a dependency (ADR 0003, 0004).

- The layout template is a Google Doc in `spilno/<association>/templates/`, created by spilno at association setup so the `drive.file` scope reaches it. spilno stores its Drive revision id; an edit made in Drive takes effect only after a член правління adopts the new revision in spilno (recorded with author and time).
- The transition to готовий до підпису commits state, reserved number (ADR 0021) and a `print_copies` row in state `pending` for that revision, and enqueues `render_pdf` (ADR 0015) in the same transaction. The worker copies the template, fills placeholders and agenda rows with values formatted by spilno's renderer (ADR 0017), exports PDF (`files.export`), stores it in the protocol folder, records Drive file id and sha256, sets the row `issued` and deletes the intermediate Doc. A PDF is rendered once and then only read; export is not byte-deterministic, so it is never regenerated for comparison.
- A return to чернетка sets the copy `void` (AC-26); withdrawal sets it `never_signed` (AC-34c); both rows are kept.

### Consequences

- Good: no new runtime dependency; layout changes need no release.
- Bad: rendering depends on Docs and Drive availability and quotas; the 5 s target becomes a job latency, measured from enqueue to `issued` (SAD §7, §10).
- Bad: variable-length tables through Docs `batchUpdate` are verbose; tests use a fake `Renderer` port and one smoke test against a test Google account.
- Neutral: a Typst adapter can replace it behind the same port.

### Migration and rollback

- Migration: new table `print_copies` (association id, protocol id, revision, number, template revision, drive file id, sha256, state, rendered at); association setting `print_template_revision`.
- Rollback: issued and void copies are kept; switching engines changes only new copies.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-06; AC-25, AC-25b, AC-26, AC-34c, AC-45; §6 "Print copy (PDF) rendering"
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §6, §7
- Related ADR: replaces the print side of [0006 Edit wording in Google Docs](./0006-edit-protocol-wording-in-google-docs-and-freeze-it-into-postgres-on-transition.md) together with [0017 Text as fragments](./0017-store-protocol-text-as-fragments-with-slot-placeholders-edited-in-spilno.md) · [0021 Protocol number](./0021-reserve-the-protocol-number-on-the-first-print-copy-and-never-reuse-it.md) · idea from [other SAD ADR 0013](../adr/0013-render-signing-copies-from-google-docs-templates.md)
