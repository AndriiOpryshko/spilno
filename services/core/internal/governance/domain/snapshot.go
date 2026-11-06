package domain

import (
	"errors"
	"time"
)

// SnapshotKind tells whether a snapshot freezes a signed or a withdrawn protocol.
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
func (s Snapshot) Canonical() ([]byte, error) {
	return nil, errors.New("not implemented")
}

// Fingerprint returns the lowercase hex SHA-256 of Canonical.
func (s Snapshot) Fingerprint() (string, error) {
	return "", errors.New("not implemented")
}
