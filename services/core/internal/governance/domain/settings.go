package domain

import (
	"strings"
	"unicode/utf8"
)

// AssociationSettings are the association-level facts a protocol needs.
type AssociationSettings struct {
	MeetingForm         string
	Signers             string
	RegisterStartNumber int
}

// MissingSettings returns the pointers of unset settings, empty when complete.
func (s AssociationSettings) MissingSettings() []string {
	var out []string
	if strings.TrimSpace(s.MeetingForm) == "" {
		out = append(out, "/meeting_form")
	}
	if strings.TrimSpace(s.Signers) == "" {
		out = append(out, "/signers")
	}
	if s.RegisterStartNumber == 0 {
		out = append(out, "/register_start_number")
	}
	return out
}

// CheckComplete fails with settings_missing naming every unset setting.
func (s AssociationSettings) CheckComplete() error {
	missing := s.MissingSettings()
	if len(missing) == 0 {
		return nil
	}
	issues := make([]FieldIssue, len(missing))
	for i, p := range missing {
		issues[i] = FieldIssue{p, "is required"}
	}
	return &CodedFieldsError{Code: CodeSettingsMissing, Fields: issues}
}

// Validate names every invalid value at once.
func (s AssociationSettings) Validate() error {
	var issues []FieldIssue
	if utf8.RuneCountInString(s.MeetingForm) > 500 {
		issues = append(issues, FieldIssue{"/meeting_form", "must be at most 500 characters"})
	}
	if utf8.RuneCountInString(s.Signers) > 500 {
		issues = append(issues, FieldIssue{"/signers", "must be at most 500 characters"})
	}
	if s.RegisterStartNumber < 0 {
		issues = append(issues, FieldIssue{"/register_start_number", "must not be negative"})
	}
	if len(issues) == 0 {
		return nil
	}
	return &CodedFieldsError{Code: CodeSettingsInvalid, Fields: issues}
}
