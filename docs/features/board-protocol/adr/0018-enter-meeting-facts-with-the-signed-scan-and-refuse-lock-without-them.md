---
status: Accepted
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-10-25"
feature_size: L
stage: "04-05"
ticket: ""
---

# 0018 — Enter the meeting facts with the signed scan and refuse the lock without them

- **Status:** Accepted
- **Date:** 2026-10-25
- **Deciders:** Andrii Opryshko (Architect), with Claude during the Stage 1 rescope

## Context

In Stage 1 the vote happens on paper at the meeting: the print copy has blank fields for members present and the vote count per item (AC-14c). The system must keep those facts with the locked protocol (AC-14) and refuse the lock without members present or a vote count for every item (AC-34); handwritten corrections on the paper need a note, kept with the scan, while the frozen text stays as is (AC-15). The informal poll must never pre-fill these facts (AC-41). Reading the scan automatically is Stage 2 (AC-14d). The facts go into the immutable lock snapshot (ADR-0007), so their shape is a long-lived contract.

Blast-radius gate: irreversible (the facts are frozen into every locked snapshot) and with legitimate alternatives — 2 of 3.

## Decision drivers

- AC-14, AC-34: lock and facts are one step; missing facts leave the protocol готовий до підпису.
- AC-15: a correction note is required when corrections are reported; the paper prevails.
- AC-41: no pre-fill from the poll.
- QG-1: the facts are part of what never changes after lock.

## Considered options

1. **Facts entered in the same request as the scan; validated first; lock in one transaction.**
2. **Facts entered before the meeting** (attendance) or right after it, independently of the scan; the lock checks they exist.
3. **Facts pre-filled from the poll tally** and corrected at upload.
4. **Facts read from the scan by the model** and confirmed by a person (Stage 2 reading, AC-14d).

## Decision outcome

**Chosen:** Option 1. One multipart request carries the scan and the facts: members present (a subset of the roster in force on the meeting's Europe/Kyiv date), for each agenda item the counts for, against and abstained, the handwritten-corrections flag and, when set, a note per correction. The API validates the state (AC-40), the term (AC-16) and the facts first and names every missing fact at once (`governance.meeting_facts_missing`, `governance.correction_note_required`); only then streams the scan to Drive and, in one transaction, stores the scan metadata, the facts with who entered them and when, builds the snapshot and sets `підписаний`. The form starts empty — nothing is copied from the poll. The system checks neither quorum nor majority nor who signed (AC-14).

## Consequences

**Positive**
- A locked protocol always has its meeting facts; there is no half-locked state.
- The poll cannot leak into the record (AC-41).
- Stage 2 scan reading (option 4) can later pre-fill the same form for confirmation without changing the stored shape.

**Negative**
- A typo in the facts is frozen with the lock; it is corrected only by a new referencing decision (AC-17).
- A large upload is retried whole if the facts are rejected; the SPA validates the facts before sending the file.

**Neutral**
- Stage 2 quorum and majority checks read the same facts.

## Links

- PRD: [PRD.md](../PRD.md) AC-14, AC-14c, AC-15, AC-16, AC-34, AC-40, AC-41 (Stage 2: AC-14d)
- SAD: [sad.md](../sad.md) §5, §6
- Related ADR: [ADR-0007](./0007-lock-protocols-as-fingerprinted-immutable-snapshots.md), [ADR-0009](./0009-track-draft-lifecycle-with-states-and-revision-numbers.md), [ADR-0005](./0005-store-files-in-association-owned-google-drive-folder.md)
