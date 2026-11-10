package domain

import (
	"net/http"
	"strings"
)

// ErrorDef describes one identity problem code and the HTTP status the
// contract answers it with.
type ErrorDef struct {
	Code   string
	Status int
}

// Problem codes of the association onboarding catalogue (PRD §5.1); each is
// listed in ErrorRegistry.
const (
	CodeApplicationInvalid          = "identity.application_invalid"
	CodeSignatureNotChecked         = "identity.signature_not_checked"
	CodeExtractCodeMismatch         = "identity.extract_code_mismatch"
	CodeAssociationAlreadyConnected = "identity.association_already_connected"
	CodeApplicationAlreadyOpen      = "identity.application_already_open"
	CodeNotCurrentMember            = "identity.not_current_member"
	CodeMismatchNeedsDocument       = "identity.mismatch_needs_document"
	CodeReasonRequired              = "identity.reason_required"
	CodeApplicationNotPending       = "identity.application_not_pending"
	CodeNotOperator                 = "identity.not_operator"
	CodeOperatorOwnApplication      = "identity.operator_own_application"
	CodeAssociationNotConnected     = "identity.association_not_connected"
	CodeAssociationStateUnchanged   = "identity.association_state_unchanged"
	CodeAssociationDisabled         = "identity.association_disabled"
	CodeTermEnded                   = "identity.term_ended"
	CodeStorageAccountPersonal      = "identity.storage_account_personal"
	CodeSessionRequired             = "identity.session_required"
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

// ErrorRegistry lists every identity.* problem code of PRD §5.1, exactly once.
var ErrorRegistry = []ErrorDef{
	{CodeApplicationInvalid, http.StatusUnprocessableEntity},
	{CodeSignatureNotChecked, http.StatusConflict},
	{CodeExtractCodeMismatch, http.StatusConflict},
	{CodeAssociationAlreadyConnected, http.StatusConflict},
	{CodeApplicationAlreadyOpen, http.StatusConflict},
	{CodeNotCurrentMember, http.StatusForbidden},
	{CodeMismatchNeedsDocument, http.StatusConflict},
	{CodeReasonRequired, http.StatusUnprocessableEntity},
	{CodeApplicationNotPending, http.StatusConflict},
	{CodeNotOperator, http.StatusForbidden},
	{CodeOperatorOwnApplication, http.StatusConflict},
	{CodeAssociationNotConnected, http.StatusConflict},
	{CodeAssociationStateUnchanged, http.StatusConflict},
	{CodeAssociationDisabled, http.StatusForbidden},
	{CodeTermEnded, http.StatusConflict},
	{CodeStorageAccountPersonal, http.StatusConflict},
	{CodeSessionRequired, http.StatusUnauthorized},
}
