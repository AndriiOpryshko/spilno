# T-A · Readiness gate on «готовий до підпису» (state transition)

Source: PRD AC-21, AC-04, AC-14c, AC-33; ADR-0009 ("чернетка or готовий до голосування → готовий до підпису
(agenda complete AC-04, every proposed slot confirmed AC-21)"). Today `Protocol.MarkReadyForSigning` never
consults `Protocol.Readiness()`, so a draft with an unconfirmed slot or no agenda can be printed.

Code: `services/core/internal/governance/domain/` (`protocol.go`, `readiness.go`, `protocol_test.go`,
`readiness_test.go`). Domain only: no HTTP, SQL or app layer change.

## Acceptance criteria

- **A1** `MarkReadyForSigning` refuses while `Readiness()` is non-empty and returns exactly the error
  `EnsureReadyForSigning()` returns (`*CodedFieldsError`, code `governance.slot_unconfirmed` or
  `governance.draft_incomplete`, every blocker in `Fields`).
- **A2** A refused call changes nothing: state, revision, number, transitions are as before, and the returned
  `Effects` is zero (no round closed, no print copy voided).
- **A3** Order of refusals: the revision/lock guard first (`protocol_locked`, `transition_not_allowed` for
  withdrawn, `draft_revision_conflict`), then `transition_not_allowed` for a protocol already готовий до
  підпису, then the readiness gate, then the number check. A stale revision on an unready draft answers
  `draft_revision_conflict`, not the readiness error.
- **A4** A ready draft (agenda item with a subject, every slot confirmed) behaves as before: from чернетка and
  from готовий до голосування (AC-33, closes the round), number taken only on the first print.
- **A5** Every existing test keeps its intent; tests that need a ready draft get one through a fixture, not by
  deleting assertions. `go test ./internal/governance/...` (unit) is green.
