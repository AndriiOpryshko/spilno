package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SnapshotKind tells whether a snapshot freezes a signed or a withdrawn protocol.
type SnapshotKind string

// Snapshot kinds (board-protocol ADR-0007).
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
	if err := s.validate(); err != nil {
		return nil, err
	}
	payload, err := canonicalValue(s.Payload, "payload")
	if err != nil {
		return nil, err
	}
	var number any
	if s.Number != 0 {
		number = s.Number
	}
	doc := map[string]any{
		"kind":           string(s.Kind),
		"meeting_date":   s.MeetingDate.Format(time.DateOnly),
		"number":         number,
		"payload":        payload,
		"protocol_id":    s.ProtocolID,
		"schema_version": s.SchemaVersion,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("snapshot: %w", err)
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// Fingerprint returns the lowercase hex SHA-256 of Canonical.
func (s Snapshot) Fingerprint() (string, error) {
	b, err := s.Canonical()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// validate checks the header fields; the payload is checked while it is normalized.
func (s Snapshot) validate() error {
	switch {
	case s.SchemaVersion < 1:
		return errors.New("snapshot: schema version must be at least 1")
	case s.Kind != SnapshotSigned && s.Kind != SnapshotWithdrawn:
		return fmt.Errorf("snapshot: unknown kind %q", s.Kind)
	case s.ProtocolID == "":
		return errors.New("snapshot: empty protocol id")
	}
	return nil
}

// canonicalValue checks v against the allowed payload types and turns every
// time.Time into its UTC RFC 3339 string; encoding/json sorts map keys.
func canonicalValue(v any, path string) (any, error) {
	switch x := v.(type) {
	case nil, string, bool, int, int64:
		return x, nil
	case time.Time:
		return x.UTC().Format(time.RFC3339Nano), nil
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			c, err := canonicalValue(e, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			c, err := canonicalValue(e, path+"."+k)
			if err != nil {
				return nil, err
			}
			out[k] = c
		}
		return out, nil
	default:
		return nil, fmt.Errorf("snapshot: unsupported type %T at %s", v, path)
	}
}
