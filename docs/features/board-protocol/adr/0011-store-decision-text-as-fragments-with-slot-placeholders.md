---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0011 — Store decision text as fragments with slot placeholders

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the architecture-design pass

> **Amended 2026-10-25 (Stage 1 rescope):** the text is edited in spilno by any current член правління (AC-29) with a structured editor; Google Docs only renders the print copy and is never an editing surface ([ADR-0013](./0013-render-signing-copies-from-google-docs-templates.md)). Шаблони формулювання (AC-23, AC-23b) are Stage 2; the same shape serves them then. The decision is unchanged.

> **Amended 2026-10-25 (cross-check with the ao-architecture SAD):** a number or fact a **person** types into a plain fragment is saved as typed and only **highlighted** in Stage 1 (with an offer to turn it into a slot); refusing or blocking it is Stage 2. The validator still turns a fact in text **written by the помічник протоколу** into a proposed, unconfirmed slot (AC-28), so the one Stage 1 print gate stays: a value the AI proposed that no person has confirmed (AC-21). Meeting facts: the meeting date and головуючий come from the meeting record; members present and the vote counts are entered from the signed paper with the scan ([ADR-0018](./0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md)) and are blank on the print copy. The decision is unchanged.

## Context

The помічник протоколу writes and rewrites the connecting text of decisions (AC-26, AC-28). That text may hold no amount, date, document reference, name or count that is not a слот рішення or a meeting fact (meeting date and головуючий from the meeting record; members present and vote counts entered from the paper with the scan); a requested change that would add one becomes a new slot to confirm (AC-28). Case documents may carry text meant to steer the proposal (PRD §6.1 abuse case "Instructions hidden in a document"). Шаблони формулювання fix wording whose slots alone may be filled (AC-23). Any number in connecting prose must equal a slot value (CONTEXT "слот рішення").

## Decision drivers

- AC-28: no fact outside slots in text the помічник протоколу writes, enforced deterministically.
- Stage 1 rescope: the only print gate is an AI-proposed value not yet confirmed by a person (AC-21); a person's own wording is not blocked.
- CONTEXT: a number in text must equal its slot value — a slot change must not leave stale text.
- AC-29: any current член правління rewords the text in spilno; every slot keeps its value and source.
- Stage 2, AC-23, AC-23b: template text fixed, only slots editable, template versions effective-dated.
- PRD §6.1: generated text holds no fact outside the slots even under prompt injection.

## Considered options

1. **Fragments with placeholders** — text is a sequence of plain-text fragments and references to slots or meeting facts; values are substituted at render time; a validator rejects facts inside plain fragments.
2. **Free text checked afterwards** — the model writes ordinary text with values inlined; the server scans it with patterns and compares with slots.

## Decision outcome

**Chosen:** Option 1. A slot change updates every place the value appears, and "no fact outside slots" is a property of the data shape plus a validator, not of how well a regular expression reads Ukrainian numerals and dates.

Shape: decision text = `[{text: "…"} | {slot: <slot id>} | {meeting: date | chair | present}]`; `date` and `chair` render from the meeting record, `present` renders blank on the print copy and from the meeting facts once they are entered with the scan. The model is asked for this structure (ADR-0004). A deterministic validator in `governance/domain` rejects a plain fragment containing digits, Ukrainian number words, month names or date forms, currency marks, document-number patterns, or any name from the roster or the case's counterparties; a rejected fact from a conversation becomes a proposed new slot, unconfirmed. Values render through one formatter per slot type (amounts in digits and in words, dates in the protocol's format), the same for screen, the copy for signing and the export. A hand edit in spilno saves the same structure with its author and time (AC-29); a fact a person types into a plain fragment is kept as typed and highlighted with an offer to make it a slot — it does not become an unconfirmed slot and does not stop printing (blocking it is Stage 2). In Stage 2 a шаблон формулювання is the same structure with its text fragments marked fixed (AC-23).

## Consequences

**Positive**
- Text the помічник протоколу writes and the numbers cannot diverge; prompt-injected facts cannot enter plain text.
- Templates and generated text share one representation and one renderer.

**Negative**
- Hand editing needs a structured editor in the SPA (text with slot chips), not a plain textarea.
- The validator's word lists (number words, month names) are Ukrainian-specific and need their own tests; a false positive forces the assistant to rephrase, or shows a needless highlight to a person.
- In Stage 1 a person can still type a number into prose that differs from its slot; the highlight shows it, nothing stops it until Stage 2.

**Neutral**
- Plain prose without facts (e.g. "to approve the works described") stays free to write.

## Links

- PRD: [PRD.md](../PRD.md) §6.1, AC-26, AC-28, AC-28b, AC-29 (Stage 2: AC-23, AC-23b)
- SAD: [sad.md](../sad.md) §5, §8
- Related ADR: [ADR-0004](./0004-use-gemini-api-for-document-reading-and-drafting.md), [ADR-0013](./0013-render-signing-copies-from-google-docs-templates.md)
