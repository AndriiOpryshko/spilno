---
id: T19b
title: "Serialize the protocol snapshot canonically and fingerprint it"
layer: "domain"
parent: T19
acs: ["AC-17"]
scope: governance
package: services/core/internal/governance/domain
test_cmd: "cd services/core && go test -count=1 ./internal/governance/domain/"
stub: services/core/internal/governance/domain/snapshot.go
test_file: services/core/internal/governance/domain/snapshot_test.go
estimate: "60-90 min"
status: "todo"
---

# T19b — Serialize the protocol snapshot canonically and fingerprint it

Slice of [T19](./T19-meeting-facts-snapshot-domain.md); [ADR-0007](../adr/0007-lock-protocols-as-fingerprinted-immutable-snapshots.md):
at lock or withdrawal the protocol is written as one canonical JSON snapshot with the SHA-256 of that
serialization; the nightly job recomputes it, so equal content must give equal bytes on every run (AC-17, QG-1).

## Interface

```go
type SnapshotKind string

const (
    SnapshotSigned    SnapshotKind = "signed"
    SnapshotWithdrawn SnapshotKind = "withdrawn"
)

// Snapshot is the frozen record of a locked or withdrawn protocol.
type Snapshot struct {
    SchemaVersion int
    Kind          SnapshotKind
    ProtocolID    string
    Number        int       // 0 = no number reserved
    MeetingDate   time.Time // a calendar date
    Payload       map[string]any
}

// Canonical returns the canonical serialization.
func (s Snapshot) Canonical() ([]byte, error)

// Fingerprint returns the lowercase hex SHA-256 of Canonical.
func (s Snapshot) Fingerprint() (string, error)
```

## Rules

- Canonical form: JSON object with keys `kind`, `meeting_date`, `number`, `payload`, `protocol_id`,
  `schema_version`; keys sorted at every nesting level; no insignificant whitespace; no HTML escaping.
- `number` is `null` when 0. `meeting_date` is `YYYY-MM-DD`.
- Payload values may be `string`, `bool`, `int`, `int64`, `nil`, `time.Time`, `[]any`, `map[string]any`.
  Integers are written as integers (amounts are kopiyky); a `time.Time` is RFC 3339 in UTC. Any other type,
  `float64` included, is an error (an amount must never be a float).
- Errors: `SchemaVersion < 1`, an unknown kind, an empty `ProtocolID`, an unsupported payload type.

## Acceptance criteria

- **AC-1** Given a signed snapshot with number 14, meeting date 2026-10-28 and payload `{"b": 1, "a": {"y": true,
  "x": "текст"}}`, Then `Canonical` is exactly
  `{"kind":"signed","meeting_date":"2026-10-28","number":14,"payload":{"a":{"x":"текст","y":true},"b":1},"protocol_id":"p-1","schema_version":1}`.
- **AC-2** Given two snapshots with equal content whose payload maps were built in different insertion orders,
  Then their `Canonical` bytes and `Fingerprint` are equal, and `Fingerprint` is the hex SHA-256 of `Canonical`.
- **AC-3** Given a snapshot, When any one of number, kind, a nested payload value or a list order changes, Then the
  fingerprint changes.
- **AC-4** Given a withdrawn snapshot with number 0, Then `number` is `null`; a payload `int64(150000000)` is written
  as `150000000` and a `time.Time` in +02:00 as its UTC RFC 3339 form.
- **AC-5** Given schema version 0, kind `"draft"`, an empty protocol id, or a `float64` anywhere in the payload,
  Then `Canonical` and `Fingerprint` return an error.
