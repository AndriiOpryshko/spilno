package domain

import "net/http"

// ErrorDef describes one governance problem code and the HTTP status the
// contract answers it with.
type ErrorDef struct {
	Code   string
	Status int
}

// ErrorRegistry lists every governance.* problem code the API contract
// advertises, exactly once.
var ErrorRegistry = []ErrorDef{
	{"governance.settings_missing", http.StatusConflict},
	{"governance.settings_invalid", http.StatusUnprocessableEntity},
	{"governance.roster_entry_invalid", http.StatusUnprocessableEntity},
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
