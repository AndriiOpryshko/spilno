package domain

import (
	"strings"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func find(ws []ReviewWarning, k WarningKind) *ReviewWarning {
	for i := range ws {
		if ws[i].Kind == k {
			return &ws[i]
		}
	}
	return nil
}

const person = "Петренко Олена Іванівна"

func TestReviewWarnings_NameMismatch(t *testing.T) {
	filing := day(2026, time.November, 10)
	fresh := filing.AddDate(0, 0, -1)
	tests := []struct {
		name         string
		typed        string
		signer, head string
		wantWarning  bool
		wantInDetail []string
		notInDetail  []string
	}{
		{"all same after levelling", "петренко  ОЛЕНА іванівна", "Олена Іванівна Петренко", person, false, nil, nil},
		{"signer has another surname", person, "Шевченко Олена Іванівна", person, true, []string{"Шевченко Олена Іванівна"}, nil},
		{"head missing patronymic", person, person, "Петренко Олена", true, []string{"Петренко Олена"}, nil},
		{"typed name in Latin", "Petrenko Olena Ivanivna", person, person, true, []string{"Petrenko Olena Ivanivna"}, nil},
		{"all three differ", "Коваль Іван Петрович", "Шевченко Олена Іванівна", "Мельник Петро Олегович",
			true, []string{"Коваль Іван Петрович", "Шевченко Олена Іванівна", "Мельник Петро Олегович"}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ws := ReviewWarnings(tc.typed, ExtractCheck{SignerName: tc.signer, HeadFullName: tc.head, ExtractDated: fresh}, filing)
			w := find(ws, WarningNameMismatch)
			if (w != nil) != tc.wantWarning {
				t.Fatalf("name_mismatch warning present = %v, want %v (warnings: %+v)", w != nil, tc.wantWarning, ws)
			}
			if w == nil {
				return
			}
			for _, s := range tc.wantInDetail {
				if !strings.Contains(w.Detail, s) {
					t.Errorf("detail %q does not name %q", w.Detail, s)
				}
			}
		})
	}
}

func TestReviewWarnings_NameMismatchNamesOnlyDifferingValues(t *testing.T) {
	filing := day(2026, time.November, 10)
	check := ExtractCheck{SignerName: "Шевченко Олена Іванівна", HeadFullName: person, ExtractDated: filing}
	w := find(ReviewWarnings(person, check, filing), WarningNameMismatch)
	if w == nil {
		t.Fatal("want a name_mismatch warning, got none")
	}
	if !strings.Contains(w.Detail, "Шевченко Олена Іванівна") {
		t.Errorf("detail %q does not name the differing signer", w.Detail)
	}
}

func TestReviewWarnings_ExtractAge(t *testing.T) {
	filing := day(2026, time.November, 10)
	tests := []struct {
		name     string
		daysOld  int
		wantWarn bool
	}{
		{"same day", 0, false},
		{"29 days", 29, false},
		{"30 days", 30, false},
		{"31 days", 31, true},
		{"200 days", 200, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dated := filing.AddDate(0, 0, -tc.daysOld)
			ws := ReviewWarnings(person, ExtractCheck{SignerName: person, HeadFullName: person, ExtractDated: dated}, filing)
			w := find(ws, WarningExtractAge)
			if (w != nil) != tc.wantWarn {
				t.Fatalf("extract_age warning present = %v, want %v (warnings: %+v)", w != nil, tc.wantWarn, ws)
			}
			if w != nil && !strings.Contains(w.Detail, dated.Format("2006-01-02")) {
				t.Errorf("detail %q does not carry the extract date %s", w.Detail, dated.Format("2006-01-02"))
			}
			if find(ws, WarningNameMismatch) != nil {
				t.Errorf("unexpected name_mismatch for identical names: %+v", ws)
			}
		})
	}
}

func TestReviewWarningKindsMatchContract(t *testing.T) {
	if WarningNameMismatch != "name_mismatch" || WarningExtractAge != "extract_age" {
		t.Errorf("kinds = %q, %q; contract ReviewWarningKind is name_mismatch, extract_age", WarningNameMismatch, WarningExtractAge)
	}
}
