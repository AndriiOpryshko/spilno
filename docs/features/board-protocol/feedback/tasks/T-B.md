# T-B · Slot confirmation goes through the protocol revision (optimistic concurrency, data integrity)

Source: ADR-0009 ("`revision` is an integer incremented by every change to text, slots, agenda items … Every
mutating request carries the revision it was built on; a mismatch answers 409 `governance.draft_revision_conflict`"),
PRD AC-21, AC-14b, AC-05d, AC-17. Today `Slot.Confirm` and `Slot.Correct` are reachable only on the slot, so
a confirmation bypasses the revision guard, the lock of a signed protocol, and the return to draft.

Code: `services/core/internal/governance/domain/` (`readiness.go` or a new `slot_edit.go`, tests next to it).
Domain only.

## Acceptance criteria

- **B1** `Protocol.ConfirmSlot(expectedRevision int, item int, key string, actorID string, at time.Time) (Effects, error)`
  and `Protocol.CorrectSlot(expectedRevision int, item int, key string, v SlotValue, actorID string, at time.Time) (Effects, error)`
  confirm / correct the slot `key` of agenda item `item` (0-based) through `Slot.Confirm` / `Slot.Correct`.
- **B2** Guard first, as every protocol mutation: signed → `governance.protocol_locked`, withdrawn →
  `governance.transition_not_allowed`, `expectedRevision != Revision()` → `governance.draft_revision_conflict`
  (stale or future). A refused call leaves the slot (status, value, origin, confirmedBy) and the protocol
  (state, revision, transitions) unchanged.
- **B3** An unknown item index (negative or out of range) or slot key is refused with a non-nil error and
  changes nothing; it must not panic.
- **B4** Success is a content change: revision +1; a protocol готовий до голосування or готовий до підпису
  returns to чернетка with the matching `Effects` (`ClosePollRound` / `VoidPrintCopy`, AC-05d, AC-14b) and a
  recorded transition; a draft stays a draft with no transition. The number is kept.
- **B5** Tests cover each of B2-B4, including a stale revision leaving a needs-review slot unconfirmed.
  `go test ./internal/governance/...` (unit) is green.
