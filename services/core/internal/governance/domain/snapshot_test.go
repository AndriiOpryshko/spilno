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
			"b": 1,
			"a": map[string]any{"y": true, "x": "текст"},
		},
	}
}

func mustCanonical(t *testing.T, s Snapshot) string {
	t.Helper()
	b, err := s.Canonical()
	if err != nil {
		t.Fatalf("Canonical: unexpected error %v", err)
	}
	return string(b)
}

func mustFingerprint(t *testing.T, s Snapshot) string {
	t.Helper()
	f, err := s.Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint: unexpected error %v", err)
	}
	return f
}

// AC-1: Given a signed snapshot with number 14, meeting date 2026-10-28 and payload
// {"b": 1, "a": {"y": true, "x": "текст"}}, Then Canonical is exactly
// {"kind":"signed","meeting_date":"2026-10-28","number":14,"payload":{"a":{"x":"текст","y":true},"b":1},"protocol_id":"p-1","schema_version":1}.
func TestSnapshot_AC1_CanonicalIsExact(t *testing.T) {
	want := `{"kind":"signed","meeting_date":"2026-10-28","number":14,"payload":{"a":{"x":"текст","y":true},"b":1},"protocol_id":"p-1","schema_version":1}`
	if got := mustCanonical(t, baseSnapshot()); got != want {
		t.Fatalf("Canonical =\n%s\nwant\n%s", got, want)
	}
}

// AC-2: Given two snapshots with equal content whose payload maps were built in different
// insertion orders, Then their Canonical bytes and Fingerprint are equal, and Fingerprint is
// the hex SHA-256 of Canonical.
func TestSnapshot_AC2_EqualContentGivesEqualBytesAndFingerprint(t *testing.T) {
	build := func(keys []string) Snapshot {
		s := baseSnapshot()
		p := map[string]any{}
		inner := map[string]any{}
		for _, k := range keys {
			p[k] = []any{"v-" + k, int64(len(k))}
			inner[k] = k
		}
		p["nested"] = inner
		s.Payload = p
		return s
	}
	a := build([]string{"alpha", "beta", "gamma", "delta", "epsilon"})
	b := build([]string{"epsilon", "delta", "gamma", "beta", "alpha"})

	ca, cb := mustCanonical(t, a), mustCanonical(t, b)
	if ca != cb {
		t.Fatalf("Canonical differs:\n%s\n%s", ca, cb)
	}
	fa, fb := mustFingerprint(t, a), mustFingerprint(t, b)
	if fa != fb {
		t.Fatalf("Fingerprint differs: %s vs %s", fa, fb)
	}
	sum := sha256.Sum256([]byte(ca))
	if want := hex.EncodeToString(sum[:]); fa != want {
		t.Fatalf("Fingerprint = %s, want sha256(Canonical) = %s", fa, want)
	}
}

// AC-3: Given a snapshot, When any one of number, kind, a nested payload value or a list order
// changes, Then the fingerprint changes.
func TestSnapshot_AC3_AnyChangeChangesFingerprint(t *testing.T) {
	withList := func() Snapshot {
		s := baseSnapshot()
		s.Payload["list"] = []any{"first", "second"}
		return s
	}
	base := mustFingerprint(t, withList())

	cases := []struct {
		name   string
		mutate func(s *Snapshot)
	}{
		{"number", func(s *Snapshot) { s.Number = 15 }},
		{"kind", func(s *Snapshot) { s.Kind = SnapshotWithdrawn }},
		{"nested payload value", func(s *Snapshot) {
			s.Payload["a"] = map[string]any{"y": false, "x": "текст"}
		}},
		{"list order", func(s *Snapshot) { s.Payload["list"] = []any{"second", "first"} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := withList()
			tc.mutate(&s)
			if got := mustFingerprint(t, s); got == base {
				t.Fatalf("fingerprint unchanged after %s change: %s", tc.name, got)
			}
		})
	}
}

// AC-4: Given a withdrawn snapshot with number 0, Then number is null; a payload
// int64(150000000) is written as 150000000 and a time.Time in +02:00 as its UTC RFC 3339 form.
func TestSnapshot_AC4_NullNumberIntegersAndUTCTime(t *testing.T) {
	s := baseSnapshot()
	s.Kind = SnapshotWithdrawn
	s.Number = 0
	s.Payload = map[string]any{
		"amount": int64(150000000),
		"at":     time.Date(2026, 10, 28, 14, 30, 0, 0, time.FixedZone("EET", 2*60*60)),
	}
	got := mustCanonical(t, s)
	for _, want := range []string{
		`"kind":"withdrawn"`,
		`"number":null`,
		`"amount":150000000`,
		`"at":"2026-10-28T12:30:00Z"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Canonical %s does not contain %s", got, want)
		}
	}
	if strings.Contains(got, "1.5e+08") {
		t.Errorf("Canonical %s writes the amount as a float", got)
	}
}

// AC-5: Given schema version 0, kind "draft", an empty protocol id, or a float64 anywhere in
// the payload, Then Canonical and Fingerprint return an error.
func TestSnapshot_AC5_InvalidSnapshotIsAnError(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(s *Snapshot)
	}{
		{"schema version 0", func(s *Snapshot) { s.SchemaVersion = 0 }},
		{"kind draft", func(s *Snapshot) { s.Kind = "draft" }},
		{"empty protocol id", func(s *Snapshot) { s.ProtocolID = "" }},
		{"top-level float64", func(s *Snapshot) { s.Payload["amount"] = float64(1500) }},
		{"nested float64", func(s *Snapshot) {
			s.Payload["a"] = map[string]any{"x": "текст", "y": true, "z": float64(1.5)}
		}},
		{"float64 in list", func(s *Snapshot) { s.Payload["list"] = []any{"ok", float64(2)} }},
	}
	// The valid base must succeed, otherwise an always-failing implementation passes this test.
	mustCanonical(t, baseSnapshot())
	mustFingerprint(t, baseSnapshot())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := baseSnapshot()
			tc.mutate(&s)
			if _, err := s.Canonical(); err == nil {
				t.Errorf("Canonical: want an error for %s", tc.name)
			}
			if _, err := s.Fingerprint(); err == nil {
				t.Errorf("Fingerprint: want an error for %s", tc.name)
			}
		})
	}
}
