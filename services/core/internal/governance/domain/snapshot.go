package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	if err := s.validate(); err != nil {
		return nil, err
	}
	var number any
	if s.Number != 0 {
		number = s.Number
	}
	doc := map[string]any{
		"kind":           string(s.Kind),
		"meeting_date":   s.MeetingDate.Format("2006-01-02"),
		"number":         number,
		"payload":        s.Payload,
		"protocol_id":    s.ProtocolID,
		"schema_version": s.SchemaVersion,
	}
	var buf bytes.Buffer
	if err := writeCanonical(&buf, doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// validate checks the snapshot header; payload types are checked while writing.
func (s Snapshot) validate() error {
	if s.SchemaVersion < 1 {
		return fmt.Errorf("snapshot: schema version %d < 1", s.SchemaVersion)
	}
	if s.Kind != SnapshotSigned && s.Kind != SnapshotWithdrawn {
		return fmt.Errorf("snapshot: unknown kind %q", s.Kind)
	}
	if s.ProtocolID == "" {
		return errors.New("snapshot: empty protocol id")
	}
	return nil
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

// writeCanonical appends v to buf as compact JSON with sorted object keys. Only
// the payload types the snapshot allows are accepted; float64 is rejected so an
// amount can never be written as a float.
func writeCanonical(buf *bytes.Buffer, v any) error {
	switch x := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		buf.WriteString(strconv.FormatBool(x))
	case int:
		buf.WriteString(strconv.Itoa(x))
	case int64:
		buf.WriteString(strconv.FormatInt(x, 10))
	case string:
		return writeString(buf, x)
	case time.Time:
		return writeString(buf, x.UTC().Format(time.RFC3339))
	case []any:
		return writeArray(buf, x)
	case map[string]any:
		return writeObject(buf, x)
	default:
		return fmt.Errorf("snapshot: unsupported payload type %T", v)
	}
	return nil
}

// writeArray appends a JSON array, keeping the element order.
func writeArray(buf *bytes.Buffer, items []any) error {
	buf.WriteByte('[')
	for i, e := range items {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeCanonical(buf, e); err != nil {
			return err
		}
	}
	buf.WriteByte(']')
	return nil
}

// writeObject appends a JSON object with its keys in sorted order.
func writeObject(buf *bytes.Buffer, obj map[string]any) error {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeString(buf, k); err != nil {
			return err
		}
		buf.WriteByte(':')
		if err := writeCanonical(buf, obj[k]); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

// writeString appends s as a JSON string without HTML escaping.
func writeString(buf *bytes.Buffer, s string) error {
	var tmp bytes.Buffer
	enc := json.NewEncoder(&tmp)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return err
	}
	buf.Write(bytes.TrimSuffix(tmp.Bytes(), []byte("\n")))
	return nil
}
