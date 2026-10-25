---
status: accepted
date: 2026-10-25
feature: board-protocol
sad_section: "§4"
deciders: ["Andrii Opryshko"]
---

# 0005 Call Gemini through a governance TextGenerator port for wording and proposals only

## Context and problem statement

The помічник протоколу writes wording and попередження, proposes amounts read from case documents (AC-03, AC-06); reading the number and date from a signed scan (AC-28b) is Stage 2. Case documents contain co-owners' names, so PRD §6.1 allows a text-generation service only under an agreement that excludes training on the data. The user chose Gemini. The Gemini API does not open Drive files by id, so the worker downloads them from Drive (ADR 0004) and sends them in the request.

## Decision drivers

- CONTEXT "помічник протоколу": every number it proposes counts only after a person confirms it.
- PRD §6: draft ≤ 30 s p95; PRD §7: ≤ 5 % of AI-proposed amounts corrected.
- PRD §6.1: no training on the association's data.
- ADR 0001 / 0015: AI calls run only in the worker.

## Blast radius

- Irreversible: no — a different provider is a new adapter behind the same port.
- Multi-context: yes — `core` enqueues, `worker` calls; web shows origin `ai` on slots.
- Legitimate alternative: yes — Claude API, or a self-hosted open-weight model.
- Touches the database: no — slots and proposals are stored the same way for any provider.

## Considered options

1. **Gemini API, paid tier, behind a `TextGenerator` port** — files sent inline or through the Files API; structured output.
2. **Claude API behind the same port.**
3. **Self-hosted open-weight model** — data stays in the hosting; needs a GPU host, weaker Ukrainian legal wording.

## Decision outcome

**Chosen:** option 1. `governance/ports.TextGenerator` has `Draft(ctx, DraftInput) (DraftProposal, error)` (Stage 2 adds `ReadSignedScan(ctx, File) (ScanReading, error)`); the adapter `governance/adapters/gemini/` uses the Google Gen AI Go SDK with a response schema (JSON shape the model must follow). The model name, API key and project come from the environment; the project must have billing enabled, because only paid Gemini API services exclude prompts from product improvement. A depguard rule `gemini-sdk` allows `google.golang.org/genai` only in `internal/governance/adapters/gemini/**`, so no other package can call the model. A `DraftProposal` holds wording per agenda item as fragments with slot placeholders (ADR 0017), попередження with a source link, and proposed slot values each with the document id and the fragment it was read from. The worker stores the wording as a new revision (ADR 0017) and the slots with origin `ai`, status unconfirmed; the transition to готовий до підпису is refused while any of them is unconfirmed (AC-06, `governance.amount_unconfirmed`).

### Consequences

- Good: the AI never confirms a number; a wrong proposal costs a correction, measured by the PRD §7 KPI.
- Good: the port keeps the provider replaceable.
- Bad: case documents leave the association's hosting; Files API copies live on Google's side up to 48 h.
- Bad: a free-tier key would silently break PRD §6.1; this is an operational check in SAD §11.

### Migration and rollback

N/A — no stored data changes beyond the slot origin already in the protocol model.

## Links

- PRD: [PRD.ao-prd.md](../PRD.ao-prd.md) — US-02, US-03, AC-03, AC-06, AC-43 (AC-28b in Stage 2); §6.1 abuse case "leaking personal data"
- SAD: [sad.ao-architecture.md](../sad.ao-architecture.md) §4, §5, §8
- Related ADR: [0004 Files on Google Drive](./0004-store-case-files-and-signed-scans-on-the-associations-google-drive-with-hashes-in-postgres.md) · [0015 River jobs](./0015-run-background-jobs-in-cmd-worker-with-river-on-postgres.md) · [0017 Text as fragments](./0017-store-protocol-text-as-fragments-with-slot-placeholders-edited-in-spilno.md)
