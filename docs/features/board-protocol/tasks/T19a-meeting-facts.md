---
id: T19a
title: "Validate the meeting facts entered with the signed scan"
layer: "domain"
parent: T19
acs: ["AC-14", "AC-34", "AC-15"]
scope: governance
package: services/core/internal/governance/domain
test_cmd: "cd services/core && go test -count=1 ./internal/governance/domain/"
stub: services/core/internal/governance/domain/meeting_facts.go
test_file: services/core/internal/governance/domain/meeting_facts_test.go
estimate: "60-90 min"
status: "done"
---

# T19a — Validate the meeting facts entered with the signed scan

Slice of [T19](./T19-meeting-facts-snapshot-domain.md); [ADR-0018](../adr/0018-enter-meeting-facts-with-the-signed-scan-and-refuse-lock-without-them.md):
the facts written on the signed paper are entered with the scan, never pre-filled from the poll, and the lock is
refused without them.

## Interface

```go
// VoteCount is the tally written on the paper for one agenda item.
type VoteCount struct{ For, Against, Abstained int }

// MeetingFacts are the facts written on the signed paper (AC-14).
type MeetingFacts struct {
    MembersPresent int
    Votes          map[int]VoteCount // keyed by agenda item index, 0-based
    HasCorrections bool
    CorrectionNote string
}

// Validate reports every missing fact at once for a protocol with agendaItems items.
func (f MeetingFacts) Validate(agendaItems int) error

// SignWithFacts locks a printed protocol with the meeting facts entered by actorID at at.
func (p *Protocol) SignWithFacts(expectedRevision int, f MeetingFacts, actorID string, at time.Time) (Effects, error)

// MeetingFacts returns the facts stored at lock, who entered them and when; ok is false before lock.
func (p *Protocol) MeetingFacts() (f MeetingFacts, enteredBy string, enteredAt time.Time, ok bool)
```

## Rules

- Errors are `*CodedFieldsError` (see `errors.go`) with JSON pointers: `/members_present`, `/votes/<i>`,
  `/correction_note`.
- Missing facts use `governance.meeting_facts_missing`; a missing note alone uses
  `governance.correction_note_required`; when both kinds are missing the code is `governance.meeting_facts_missing`
  and every pointer is listed. Pointers are ordered: members present, votes by item index, correction note.
- Quorum, majority and who signed are not checked (Stage 1, AC-14).

## Acceptance criteria

- **AC-1** Given 3 agenda items, When the facts have members present 5, a vote count for items 0, 1 and 2 and no
  corrections, Then `Validate(3)` returns nil.
- **AC-2** Given 3 agenda items, When members present is 0 and the vote counts for items 0 and 2 are missing,
  Then `Validate(3)` returns `governance.meeting_facts_missing` naming `/members_present`, `/votes/0`, `/votes/2`
  in that order (every missing fact at once, AC-34).
- **AC-3** Given complete facts with corrections reported, When the note is empty or only whitespace, Then
  `Validate` returns `governance.correction_note_required` naming `/correction_note` (AC-15); with a note it is nil.
- **AC-4** Given members present 0 and corrections reported without a note, Then the code is
  `governance.meeting_facts_missing` and the pointers are `/members_present`, `/correction_note`.
- **AC-5** Given a protocol готовий до підпису with 2 agenda items, When `SignWithFacts` gets incomplete facts,
  Then it returns the `Validate` error, the protocol stays готовий до підпису with the same revision and
  transitions, and `MeetingFacts()` reports ok false (AC-34).
- **AC-6** Given the same protocol, When `SignWithFacts` gets complete facts, Then the protocol is підписаний, and
  `MeetingFacts()` returns those facts with enteredBy = actor and enteredAt = at (AC-14). The usual guard applies
  first: a stale revision is `governance.draft_revision_conflict`, a draft is `governance.not_ready_for_signing`.
