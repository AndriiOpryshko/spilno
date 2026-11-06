package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"
)

func baseSnapshot() Snapshot {
	return Snapshot{
		SchemaVersion: 1,
		Kind:          SnapshotSigned,
		ProtocolID:    "p-1",
		Number:        14,
		MeetingDate:   time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC),
		Payload: map[string]any{
			"b":     1,
			"a":     map[string]any{"y": true, "x": "текст"},
			"items": []any{"first", "second"},
		},
	}
}

func mustFingerprint(t *testing.T, s Snapshot) string {
	t.Helper()
	fp, err := s.Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint: unexpected error %v", err)
	}
	return fp
}

// AC-1: keys sorted at every level, compact, no HTML escaping, date as YYYY-MM-DD.
func TestSnapshotCanonical_AC1_ExactBytes(t *testing.T) {
	s := baseSnapshot()
	s.Payload = map[string]any{"b": 1, "a": map[string]any{"y": true, "x": "текст"}}

	got, err := s.Canonical()
	if err != nil {
		t.Fatalf("Canonical: unexpected error %v", err)
	}
	want := `{"kind":"signed","meeting_date":"2026-10-28","number":14,"payload":{"a":{"x":"текст","y":true},"b":1},"protocol_id":"p-1","schema_version":1}`
	if string(got) != want {
		t.Fatalf("Canonical =\n%s\nwant\n%s", got, want)
	}
}

// AC-2: insertion order does not matter; Fingerprint is hex SHA-256 of Canonical.
func TestSnapshotFingerprint_AC2_OrderIndependentAndSHA256(t *testing.T) {
	a := baseSnapshot()
	a.Payload = map[string]any{}
	a.Payload["z"] = "last"
	a.Payload["m"] = map[string]any{"q": 2, "p": "<&>"}
	a.Payload["a"] = []any{map[string]any{"k2": false, "k1": nil}}

	b := baseSnapshot()
	b.Payload = map[string]any{}
	b.Payload["a"] = []any{map[string]any{"k1": nil, "k2": false}}
	b.Payload["m"] = map[string]any{"p": "<&>", "q": 2}
	b.Payload["z"] = "last"

	ca, err := a.Canonical()
	if err != nil {
		t.Fatalf("Canonical a: %v", err)
	}
	cb, err := b.Canonical()
	if err != nil {
		t.Fatalf("Canonical b: %v", err)
	}
	if string(ca) != string(cb) {
		t.Fatalf("Canonical differs:\n%s\n%s", ca, cb)
	}
	if !strings.Contains(string(ca), `"p":"<&>"`) {
		t.Errorf("Canonical %s: want unescaped <&>", ca)
	}
	fa, fb := mustFingerprint(t, a), mustFingerprint(t, b)
	if fa != fb {
		t.Fatalf("Fingerprint differs: %s vs %s", fa, fb)
	}
	sum := sha256.Sum256(ca)
	if want := hex.EncodeToString(sum[:]); fa != want {
		t.Fatalf("Fingerprint = %q, want sha256(Canonical) %q", fa, want)
	}
}

// AC-3: any single content change changes the fingerprint.
func TestSnapshotFingerprint_AC3_AnyChangeChangesFingerprint(t *testing.T) {
	base := mustFingerprint(t, baseSnapshot())

	cases := map[string]func(*Snapshot){
		"number": func(s *Snapshot) { s.Number = 15 },
		"kind":   func(s *Snapshot) { s.Kind = SnapshotWithdrawn },
		"nested value": func(s *Snapshot) {
			s.Payload["a"] = map[string]any{"y": false, "x": "текст"}
		},
		"list order": func(s *Snapshot) { s.Payload["items"] = []any{"second", "first"} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := baseSnapshot()
			mutate(&s)
			if got := mustFingerprint(t, s); got == base {
				t.Fatalf("fingerprint unchanged after %s change: %s", name, got)
			}
		})
	}
}

// AC-4: number 0 → null; int64 as integer; time.Time as UTC RFC 3339.
func TestSnapshotCanonical_AC4_NullNumberInt64AndUTCTime(t *testing.T) {
	s := Snapshot{
		SchemaVersion: 1,
		Kind:          SnapshotWithdrawn,
		ProtocolID:    "p-2",
		Number:        0,
		MeetingDate:   time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC),
		Payload: map[string]any{
			"amount_kopiyky": int64(150000000),
			"at":             time.Date(2026, 10, 28, 19, 30, 0, 0, time.FixedZone("EET", 2*60*60)),
		},
	}
	got, err := s.Canonical()
	if err != nil {
		t.Fatalf("Canonical: unexpected error %v", err)
	}
	want := `{"kind":"withdrawn","meeting_date":"2026-11-03","number":null,"payload":{"amount_kopiyky":150000000,"at":"2026-10-28T17:30:00Z"},"protocol_id":"p-2","schema_version":1}`
	if string(got) != want {
		t.Fatalf("Canonical =\n%s\nwant\n%s", got, want)
	}
}

// AC-5: invalid header fields or a float anywhere in the payload are errors.
func TestSnapshot_AC5_InvalidSnapshotsError(t *testing.T) {
	cases := map[string]func(*Snapshot){
		"schema version 0":  func(s *Snapshot) { s.SchemaVersion = 0 },
		"kind draft":        func(s *Snapshot) { s.Kind = "draft" },
		"empty protocol id": func(s *Snapshot) { s.ProtocolID = "" },
		"float64 top level": func(s *Snapshot) { s.Payload["amount"] = 1.5 },
		"float64 nested": func(s *Snapshot) {
			s.Payload["a"] = map[string]any{"deep": []any{"ok", float64(2)}}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := baseSnapshot()
			mutate(&s)
			if b, err := s.Canonical(); err == nil {
				t.Errorf("Canonical: want error, got nil (bytes %q)", b)
			}
			if fp, err := s.Fingerprint(); err == nil {
				t.Errorf("Fingerprint: want error, got nil (fingerprint %q)", fp)
			}
		})
	}
}
