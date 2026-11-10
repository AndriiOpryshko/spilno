package domain

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func pointersOf(t *testing.T, err error, wantCode string) []string {
	t.Helper()
	var ve *CodedFieldsError
	if !errors.As(err, &ve) {
		t.Fatalf("error = %v (%T), want *CodedFieldsError", err, err)
	}
	if ve.Code != wantCode {
		t.Errorf("code = %q, want %q", ve.Code, wantCode)
	}
	got := ve.Pointers()
	sort.Strings(got)
	return got
}

func TestBoardMemberEntry_Validate(t *testing.T) {
	valid := BoardMemberEntry{FullName: "Test User", SignInEmail: "user@example.test", Term: Term{Start: day(2026, 1, 1)}}

	tests := []struct {
		name  string
		entry BoardMemberEntry
		want  []string // nil = valid
	}{
		{"valid open term", valid, nil},
		{"valid closed term, no email", BoardMemberEntry{FullName: "Test User", Term: Term{Start: day(2025, 1, 1), End: day(2026, 1, 1)}}, nil},
		{"end equal to start is allowed", BoardMemberEntry{FullName: "Test User", Term: Term{Start: day(2026, 1, 1), End: day(2026, 1, 1)}}, nil},
		{"AC-02 missing term start", BoardMemberEntry{FullName: "Test User"}, []string{"/term_start"}},
		{"AC-02 end before start", BoardMemberEntry{FullName: "Test User", Term: Term{Start: day(2026, 5, 1), End: day(2026, 4, 30)}}, []string{"/term_end"}},
		{"AC-02 every broken field at once", BoardMemberEntry{FullName: "  ", Term: Term{End: day(2026, 4, 30)}}, []string{"/full_name", "/term_start"}},
		{"name too long and email too long", BoardMemberEntry{
			FullName: strings.Repeat("я", 201), SignInEmail: strings.Repeat("a", 320) + "@example.test",
			Term: Term{Start: day(2026, 5, 1), End: day(2026, 4, 1)},
		}, []string{"/full_name", "/sign_in_email", "/term_end"}},
		{"name of 200 runes is fine", BoardMemberEntry{FullName: strings.Repeat("я", 200), Term: Term{Start: day(2026, 1, 1)}}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.Validate()
			if tc.want == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			got := pointersOf(t, err, "governance.roster_entry_invalid")
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("broken fields = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTerm_CoversDate(t *testing.T) {
	closed := Term{Start: day(2026, 1, 1), End: day(2026, 12, 31)}
	open := Term{Start: day(2026, 1, 1)}
	tests := []struct {
		name string
		term Term
		date time.Time
		want bool
	}{
		{"before start", closed, day(2025, 12, 31), false},
		{"on start", closed, day(2026, 1, 1), true},
		{"inside", closed, day(2026, 6, 15), true},
		{"on end", closed, day(2026, 12, 31), true},
		{"after end", closed, day(2027, 1, 1), false},
		{"open: before start", open, day(2025, 12, 31), false},
		{"open: on start", open, day(2026, 1, 1), true},
		{"open: far future", open, day(2099, 1, 1), true},
		{"time of day ignored on end day", closed, time.Date(2026, 12, 31, 23, 59, 0, 0, time.UTC), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.term.CoversDate(tc.date); got != tc.want {
				t.Errorf("CoversDate(%s) = %v, want %v", tc.date.Format("2006-01-02"), got, tc.want)
			}
		})
	}
}

func TestBoardMemberEntry_Validate_Position(t *testing.T) {
	base := BoardMemberEntry{FullName: "Test User", Term: Term{Start: day(2026, 1, 1)}}

	for _, p := range []Position{PositionHead, PositionMember} {
		e := base
		e.Position = p
		e.DecisionID = "0198f0a0-0000-7000-8000-000000000001"
		if err := e.Validate(); err != nil {
			t.Errorf("position %q: Validate() = %v, want nil", p, err)
		}
	}

	e := base
	e.Position = Position("chair")
	got := pointersOf(t, e.Validate(), CodeRosterEntryInvalid)
	if want := []string{"/position"}; !reflect.DeepEqual(got, want) {
		t.Errorf("pointers = %v, want %v", got, want)
	}
}

func TestPositionValues(t *testing.T) {
	if PositionHead != "head" || PositionMember != "member" {
		t.Errorf("positions = %q, %q; want head, member", PositionHead, PositionMember)
	}
}

func TestPreviousHeadTermEnd(t *testing.T) {
	tests := []struct {
		name                       string
		newTermStart, previousFrom time.Time
		want                       time.Time
	}{
		{"day before the new term start", day(2026, 11, 20), day(2025, 1, 1), day(2026, 11, 19)},
		{"across a month boundary", day(2026, 12, 1), day(2025, 1, 1), day(2026, 11, 30)},
		{"never before the entry's first day", day(2026, 11, 20), day(2026, 11, 25), day(2026, 11, 25)},
		{"new term starts the day the entry began", day(2026, 11, 20), day(2026, 11, 20), day(2026, 11, 20)},
		{"entry began the day before the new term start", day(2026, 11, 20), day(2026, 11, 19), day(2026, 11, 19)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PreviousHeadTermEnd(tc.newTermStart, tc.previousFrom)
			if dateKey(got) != dateKey(tc.want) {
				t.Errorf("PreviousHeadTermEnd(%s, %s) = %s, want %s",
					tc.newTermStart.Format("2006-01-02"), tc.previousFrom.Format("2006-01-02"),
					got.Format("2006-01-02"), tc.want.Format("2006-01-02"))
			}
		})
	}
}

func TestReseat(t *testing.T) {
	term := Term{Start: day(2024, 3, 1), End: day(2027, 3, 1)}
	entry := BoardMemberEntry{FullName: "Test User", SignInEmail: "user@example.test", Position: PositionMember, Term: term}

	got := Reseat(entry, "0198f0a0-0000-7000-8000-000000000002")

	if got.Position != PositionHead {
		t.Errorf("Position = %q, want %q", got.Position, PositionHead)
	}
	if got.DecisionID != "0198f0a0-0000-7000-8000-000000000002" {
		t.Errorf("DecisionID = %q, want the decision reference", got.DecisionID)
	}
	if got.Term != term {
		t.Errorf("Term = %v, want the entry's own term %v", got.Term, term)
	}
	if got.FullName != entry.FullName || got.SignInEmail != entry.SignInEmail {
		t.Errorf("identity changed: %+v, want %q / %q", got, entry.FullName, entry.SignInEmail)
	}
	if entry.Position != PositionMember || entry.DecisionID != "" {
		t.Errorf("input entry mutated: %+v", entry)
	}
}
