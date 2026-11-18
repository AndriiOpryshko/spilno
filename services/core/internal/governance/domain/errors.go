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

// Problem codes raised by the protocol lifecycle; each is listed in ErrorRegistry.
const (
	CodeDraftIncomplete    = "governance.draft_incomplete"
	CodeRevisionConflict   = "governance.draft_revision_conflict"
	CodeProtocolLocked     = "governance.protocol_locked"
	CodeTransitionDenied   = "governance.transition_not_allowed"
	CodeNotReadyForSigning = "governance.not_ready_for_signing"
	CodeReasonRequired     = "governance.reason_required"
	CodePollNotOpen        = "governance.poll_not_open"
	CodeVoteOnBehalf       = "governance.vote_on_behalf"
	CodeFactsMissing       = "governance.meeting_facts_missing"
	CodeNoteRequired       = "governance.correction_note_required"
)

// Problem codes of uploads, the registry search, printing and the register start; each is listed in ErrorRegistry.
const (
	// CodeFileInvalid is answered for an upload over MaxDocumentSize or not PDF, JPEG or PNG by content.
	CodeFileInvalid = "governance.file_invalid"
	// CodeSearchInvalid is answered for a registry search the filters of which are invalid.
	CodeSearchInvalid = "governance.search_invalid"
	// CodeSlotUnconfirmed blocks printing while a слот рішення is unconfirmed or needs review (AC-21, AC-28b, AC-09).
	CodeSlotUnconfirmed = "governance.slot_unconfirmed"
)

// CodeStorageUnavailable is answered when the association's file storage
// cannot be reached; the request changes nothing (sad §6 Flow 8).
const CodeStorageUnavailable = "governance.storage_unavailable"

// CodeDriveNotConnected is answered when the association has no connected
// Google Drive storage; each is listed in ErrorRegistry.
const CodeDriveNotConnected = "governance.drive_not_connected"

// DomainError is a rule violation identified only by its problem code.
type DomainError struct {
	Code string
}

func (e *DomainError) Error() string { return e.Code }

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
	{CodeFileInvalid, http.StatusUnprocessableEntity},
	{CodeDraftIncomplete, http.StatusConflict},
	{CodeSlotUnconfirmed, http.StatusConflict},
	{CodeFactsMissing, http.StatusUnprocessableEntity},
	{CodeNoteRequired, http.StatusUnprocessableEntity},
	{CodeReasonRequired, http.StatusUnprocessableEntity},
	{CodeSearchInvalid, http.StatusUnprocessableEntity},
	{CodeVoteOnBehalf, http.StatusForbidden},
	{CodeRevisionConflict, http.StatusConflict},
	{CodeNotReadyForSigning, http.StatusConflict},
	{CodePollNotOpen, http.StatusConflict},
	{CodeProtocolLocked, http.StatusConflict},
	{"governance.register_start_locked", http.StatusConflict},
	{CodeTransitionDenied, http.StatusConflict},
	{CodeStorageUnavailable, http.StatusServiceUnavailable},
	{CodeDriveNotConnected, http.StatusConflict},
}
