---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0004 — Use the Gemini API, behind a port, to read case documents and to draft

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** drafting through the conversation and pre-filling slots moved into Stage 1; reading the signed scan (AC-14d) moved to Stage 2. The port and the only Gemini adapter live in `governance` (`governance/ports.TextModel`, `governance/adapters/gemini`), not in an `assistant` context ([ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)); a depguard rule keeps the Gemini SDK out of every other package. Targets restated against PRD §6 (p95 ≤ 30 s, files ≤ 20 MB). The choice of Gemini is unchanged.

## Context

Stage 1 drafts a protocol from a conversation (AC-26, AC-28, AC-28b) and pre-fills слоти рішення from the case documents (AC-21, AC-22). The documents are Ukrainian PDFs and scans: contracts, estimates, invoices, acts. PRD §8 allows processing outside the association's own hosting only under an agreement that excludes training on the data. The association's files live in its Google Drive ([ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)).

## Decision drivers

- PRD §6: proposed protocol p95 ≤ 30 s from request to filled draft; files up to 20 MB.
- PRD §6.1: confidential personal data in case documents; abuse case "Instructions hidden in a document".
- AC-21, AC-28, AC-28b: every proposed number stays unconfirmed until a person confirms it; generated text holds no fact outside the slots; a confirmed slot is never overwritten.
- ADR-0005: files are already in the association's Google account.

## Considered options

1. **Claude API** behind a port — Anthropic Messages API, reads PDFs and images natively, structured output.
2. **A port with both Claude and Gemini adapters, default chosen by a measured bake-off** on the historical protocols' documents.
3. **Gemini API, documents taken from the association's Google Drive** — the worker fetches files from Drive and sends them to Gemini.
4. **Self-hosted open model (Ollama)** — data stays in house; needs a GPU host and reads Ukrainian scans noticeably worse.

## Decision outcome

**Chosen:** Option 3 — Gemini API, documents from the association's Drive. It keeps the association's files and their processing with one provider whose account the association already uses, which simplifies the PRD §8 approval to one agreement. Only the **paid** tier of the Gemini API is used: its terms exclude using prompts and responses to improve the provider's products, while the free tier does not.

Shape: `governance/ports.TextModel` is the only way any code reaches a model; `governance/adapters/gemini` implements it and is the only package allowed to import the Gemini SDK (depguard `gemini-sdk-in-adapter`). The worker's `propose_draft` job downloads the file content from Drive and sends it inline; only a file larger than the inline request limit goes through the Gemini Files API, and that upload is deleted right after the response (the Files API otherwise keeps it up to 48 h — SAD §11). Requests ask for JSON matching a schema: per slot a value, the quoted fragment, the document version id and a confidence note; the connecting text as fragments with slot placeholders ([ADR-0011](./0011-store-decision-text-as-fragments-with-slot-placeholders.md)). Only the documents of the справа being drafted are sent. Responses are validated against the schema and the case's document ids; anything else is discarded. Model name and tier are configuration.

## Consequences

**Positive**
- One provider for files and processing; native PDF and image reading, no separate OCR engine.
- The port keeps a second adapter (Claude, Vertex AI, a local model) a contained change in `governance/adapters`.
- Output is schema-checked data, so "people confirm every number" is enforced by `governance`, not by trusting the model.

**Negative**
- Personal data of co-owners and contractors leaves the association's hosting for the provider; the board must approve it under the paid-tier terms (SAD §11).
- Files sent through the Files API may stay on Google's side up to 48 h unless deleted (SAD §11, accepted debt).
- No measured comparison against other models before committing; extraction quality on the association's own documents is unknown until the evaluation set runs (SAD §11).
- Requires a billing-enabled Google Cloud project and an API key kept as a secret.

**Neutral**
- Prompt texts and the evaluation set are kept with the adapter so a provider switch reuses the evaluation, not the prompts.
- Reading the number and date from a signed scan (AC-14d) reuses this port in Stage 2.

## Links

- PRD: [PRD.md](../PRD.md) §6, §6.1, §8, AC-21, AC-22, AC-26, AC-28, AC-28b
- SAD: [sad.md](../sad.md) §4, §5, §8, §11
- Related ADR: [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md), [ADR-0006](./0006-run-background-work-in-worker-via-river-queue.md), [ADR-0014](./0014-keep-board-protocol-in-governance-and-add-only-an-identity-context.md)
