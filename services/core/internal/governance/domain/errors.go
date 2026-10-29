package domain

import (
	"net/http"
	"strings"
)

// ErrorDef describes one governance problem code and the HTTP status the
// contract answers it with.
type ErrorDef struct {
	Code   string
	Status int
}

// Problem codes raised by the roster and settings rules; each is listed in ErrorRegistry.
const (
	CodeRosterEntryInvalid = "governance.roster_entry_invalid"
	CodeSettingsInvalid    = "governance.settings_invalid"
	CodeSettingsMissing    = "governance.settings_missing"
)

// FieldIssue names one broken field by JSON pointer.
type FieldIssue struct {
	Pointer string
	Detail  string
}

// CodedFieldsError carries a problem code and every broken field at once.
type CodedFieldsError struct {
	Code   string
	Fields []FieldIssue
}

func (e *CodedFieldsError) Error() string {
	return e.Code + ": " + strings.Join(e.Pointers(), ", ")
}

// Pointers returns the pointers of the broken fields in order.
func (e *CodedFieldsError) Pointers() []string {
	out := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		out[i] = f.Pointer
	}
	return out
}

// ErrorRegistry lists every governance.* problem code the API contract
// advertises, exactly once.
var ErrorRegistry = []ErrorDef{
	{CodeSettingsMissing, http.StatusConflict},
	{CodeSettingsInvalid, http.StatusUnprocessableEntity},
	{CodeRosterEntryInvalid, http.StatusUnprocessableEntity},
	{"governance.file_invalid", http.StatusUnprocessableEntity},
	{"governance.draft_incomplete", http.StatusConflict},
	{"governance.slot_unconfirmed", http.StatusConflict},
	{"governance.meeting_facts_missing", http.StatusUnprocessableEntity},
	{"governance.correction_note_required", http.StatusUnprocessableEntity},
	{"governance.reason_required", http.StatusUnprocessableEntity},
	{"governance.search_invalid", http.StatusUnprocessableEntity},
	{"governance.vote_on_behalf", http.StatusForbidden},
	{"governance.draft_revision_conflict", http.StatusConflict},
	{"governance.not_ready_for_signing", http.StatusConflict},
	{"governance.poll_not_open", http.StatusConflict},
	{"governance.protocol_locked", http.StatusConflict},
	{"governance.register_start_locked", http.StatusConflict},
	{"governance.transition_not_allowed", http.StatusConflict},
	{"governance.storage_unavailable", http.StatusServiceUnavailable},
}
