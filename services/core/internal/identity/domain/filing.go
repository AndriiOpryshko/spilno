package domain

import (
	"time"
	"unicode/utf8"
)

type FilingKind string

const (
	FilingConnection FilingKind = "connection"
	FilingTransfer   FilingKind = "transfer"
)

const PositionHead = "head"

type Filing struct {
	Kind          FilingKind
	Edrpou        string
	Name          string
	Address       string
	ApplicantName string
	Position      string
	TermStart     time.Time
	TermEnd       *time.Time
	Extract       *ExtractFile
}

// Validate names every broken field at once, then checks the term end.
// filingDate is compared by its Europe/Kyiv calendar day.
func (f Filing) Validate(filingDate time.Time) error {
	today := dayOf(filingDate.In(kyivZone()))
	var issues []FieldIssue
	add := func(p, d string) { issues = append(issues, FieldIssue{Pointer: p, Detail: d}) }

	if !isEdrpou(f.Edrpou) {
		add("/edrpou", "must be 8 digits")
	}
	if !runeLenBetween(f.Name, 1, 200) {
		add("/association_name", "must be 1 to 200 characters")
	}
	if !runeLenBetween(f.Address, 1, 500) {
		add("/association_address", "must be 1 to 500 characters")
	}
	if !runeLenBetween(f.ApplicantName, 1, 200) {
		add("/applicant_full_name", "must be 1 to 200 characters")
	}
	if f.Position != PositionHead {
		add("/position", "must be head")
	}
	if f.TermStart.IsZero() || dayOf(f.TermStart).After(today) {
		add("/term_start", "must be a date not later than the filing date")
	}
	if f.Extract == nil {
		add("/extract", "is required")
	} else if err := f.Extract.Validate(); err != nil {
		add("/extract", err.Error())
	}
	if len(issues) > 0 {
		return &CodedFieldsError{Code: CodeApplicationInvalid, Fields: issues}
	}
	if f.TermEnd != nil && dayOf(*f.TermEnd).Before(today) {
		return &CodedFieldsError{Code: CodeTermEnded, Fields: []FieldIssue{{Pointer: "/term_end", Detail: "the term ended before the filing date"}}}
	}
	return nil
}

func runeLenBetween(s string, lo, hi int) bool {
	n := utf8.RuneCountInString(s)
	return n >= lo && n <= hi
}

func isEdrpou(s string) bool {
	if len(s) != 8 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func dayOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func kyivZone() *time.Location {
	if l, err := time.LoadLocation("Europe/Kyiv"); err == nil {
		return l
	}
	return time.UTC
}
