---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0013 — Render print copies from Google Docs layout templates

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** Google Docs only renders the print copy — it is never an editing surface for the protocol text, which is edited in spilno ([ADR-0011](./0011-store-decision-text-as-fragments-with-slot-placeholders.md)). The copy carries the settings (meeting form, signers), the reserved number ([ADR-0017](./0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)) and blank fields for members present and the vote count per item (AC-14c). Under `drive.file` spilno creates the layout template itself. The export package (AC-20) is Stage 2. The choice of Docs rendering is unchanged.

## Context

Marking a draft готовий до підпису produces the PDF that is printed and signed at the meeting, with the meeting date, the meeting form and the signers from the settings, the protocol number, and blank fields for members present and the vote count per item to fill in by hand (AC-14c); it carries no poll result (AC-33, AC-41). A later edit voids it and the member is told not to sign earlier copies (AC-14b). Text must be Ukrainian with proper fonts. All files live in the association's Google Drive under the `drive.file` scope ([ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)). The author asked to reuse Google Docs for this.

## Decision drivers

- PRD §6: print copy rendering p95 ≤ 5 s.
- AC-14b, AC-14c: the system must know exactly which copy was rendered and whether it is still valid.
- ADR-0005: files in the association's Drive; no new service preferred for one developer.

## Considered options

1. **Typst on the server** — a template in the repo rendered by the `typst` binary in the worker; PDF uploaded to Drive.
2. **HTML → PDF through Gotenberg** — a headless-Chromium container rendering HTML templates.
3. **Browser print** — an SPA print page; the user saves a PDF.
4. **Google Docs templates + Drive PDF export** — the worker copies a template Doc, fills it through the Docs API, exports PDF through the Drive API.

## Decision outcome

**Chosen:** Option 4. No new binary or service, Cyrillic and fonts work as in any Google Doc, and the board adjusts the layout in an editor it knows.

Shape:
- The layout template is a Google Doc in `spilno/templates/`, created by spilno from a seed kept in the repo so that `drive.file` covers it. spilno pins the template's Drive revision: an edit made in Docs takes effect only after a член правління adopts it in spilno (recorded with author and time). Layout templates are distinct from шаблони формулювання (Stage 2).
- On готовий до підпису the worker's `render_print_copy` job copies the pinned template, fills placeholders and table rows (agenda items with the frozen text, slot values formatted by spilno's renderer — ADR-0011, meeting date as the Europe/Kyiv calendar date, meeting form, signers, number; blank attendance and vote-count rows) through Docs `batchUpdate`, exports PDF with `files.export`, stores it in `spilno/protocols/<id>/`, records the print copy (revision, number, Drive id, sha256 — ADR-0009) and deletes the intermediate Doc. The job is idempotent per protocol and revision. A PDF is rendered once and then only read: export is not byte-deterministic, so it is never regenerated for comparison.

## Consequences

**Positive**
- No new runtime dependency; files and rendering stay in the association's Google account.
- Layout changes need no release.

**Negative**
- Rendering depends on Google API availability and quotas; three API calls per copy put the 5 s target at risk (SAD §11), retried by River.
- Variable-length tables through `batchUpdate` are verbose to build and to test; integration tests need a test Google account or a recorded fake.
- Someone with Drive access can change the template; pinning and adoption keep it from reaching protocols unnoticed, and the rendered PDF's sha256 is what counts.

**Neutral**
- A Typst renderer could replace the Docs adapter behind the same `governance/ports.Renderer`.
- The Stage 2 export summary reuses the same renderer.

## Links

- PRD: [PRD.md](../PRD.md) §6, AC-14b, AC-14c, AC-33, AC-41
- SAD: [sad.md](../sad.md) §5, §6, §10, §11
- Related ADR: [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md), [ADR-0009](./0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0011](./0011-store-decision-text-as-fragments-with-slot-placeholders.md), [ADR-0017](./0017-reserve-the-protocol-number-at-first-print-and-never-reuse-it.md)
