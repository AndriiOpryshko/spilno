---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0004 — Use the Gemini API, behind a port, to read case documents and scans and to draft

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

## Context

Stage 2 pre-fills слоти рішення from case documents (AC-21, AC-22), drafts a protocol from a conversation (AC-26 … AC-28b), and Stage 1 reads the protocol number and date from an uploaded signed scan (AC-14d). The documents are Ukrainian PDFs and scans: contracts, estimates, invoices, acts. PRD §8 allows processing outside the association's own hosting only under an agreement that excludes training on the data. The association's files live in its Google Drive (ADR-0005).

## Decision drivers

- PRD §6: proposed protocol p95 ≤ 60 s for up to 10 attached documents; signed-copy upload p95 ≤ 20 s including reading the number and date.
- PRD §6.1: confidential personal data in case documents; abuse case "Instructions hidden in a document".
- AC-21, AC-27, AC-28: every proposed number stays unconfirmed until a person confirms it; generated text holds no fact outside the slots.
- ADR-0005: files are already in the association's Google account.

## Considered options

1. **Claude API** behind a port — Anthropic Messages API, reads PDFs and images natively, structured output.
2. **A port with both Claude and Gemini adapters, default chosen by a measured bake-off** on the historical protocols' documents.
3. **Gemini API, documents taken from the association's Google Drive** — the worker fetches files from Drive and sends them to Gemini.
4. **Self-hosted open model (Ollama)** — data stays in house; needs a GPU host and reads Ukrainian scans noticeably worse.

## Decision outcome

**Chosen:** Option 3 — Gemini API, documents from the association's Drive. It keeps the association's files and their processing with one provider whose account the association already uses, which simplifies the PRD §8 approval to one agreement. Only the **paid** tier of the Gemini API is used: its terms exclude using prompts and responses to improve the provider's products, while the free tier does not.

Shape: `assistant/ports.TextModel` is the only way any code reaches a model; the Gemini adapter implements it. The worker downloads the file content from Drive and sends it inline, or through the Gemini Files API when larger than the inline limit (signed scans up to 50 MB). Requests ask for JSON matching a schema: per slot a value, the quoted fragment, the document version id and a confidence note; per scan the number and date read or "unreadable". Only the documents of the справа being drafted are sent. Responses are validated against the schema and the case's document ids; anything else is discarded. Model name and tier are configuration.

## Consequences

**Positive**
- One provider for files and processing; native PDF and image reading, no separate OCR engine.
- The port keeps a second adapter (Claude, Vertex AI, a local model) a contained change in `assistant/adapters`.
- Output is schema-checked data, so "people confirm every number" is enforced by `governance`, not by trusting the model.

**Negative**
- Personal data of co-owners and contractors leaves the association's hosting for the provider; the board must approve it under the paid-tier terms (tracked in §11).
- No measured comparison against other models before committing; extraction quality on the association's own scans is unknown until the regression set runs (tracked in §11).
- Requires a billing-enabled Google Cloud project and an API key kept as a secret.

**Neutral**
- Prompt texts and the evaluation set are kept with the adapter so a provider switch reuses the evaluation, not the prompts.

## Links

- PRD: [[../PRD.md]] §6, §6.1, §8, AC-14d, AC-21, AC-22, AC-26, AC-27, AC-28, AC-28b
- SAD: [[../sad.md]] §4, §5, §11
- Related ADR: [[0005-store-files-in-association-owned-google-drive-folder]], [[0006-run-background-work-in-worker-via-river-queue]]
