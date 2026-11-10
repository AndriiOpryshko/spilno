package domain

import (
	"strings"
	"unicode/utf8"
)

type ApplicationKind string

type ApplicationState string

const (
	KindConnection ApplicationKind = "connection"
	KindTransfer   ApplicationKind = "transfer"

	StatePending  ApplicationState = "pending"
	StateApproved ApplicationState = "approved"
	StateRejected ApplicationState = "rejected"
)

// CodedError is a refused transition carrying its catalogue code.
type CodedError struct{ Code string }

func (e *CodedError) Error() string { return e.Code }

// Check is the оператор's recorded finding for the current filing.
type Check struct {
	SignatureValid bool
	ExtractEDRPOU  string
}

// Application is a заявка на підключення or a запит на передачу.
type Application struct {
	Kind      ApplicationKind
	State     ApplicationState
	AccountID string
	EDRPOU    string
	Check     *Check
}

// ApprovalInput is what the оператор brings to an approval.
type ApprovalInput struct {
	OperatorAccountID string
	AssociationEDRPOU string
	MismatchOpen      bool
	Document          string
}

// documentMaxRunes bounds the document that justifies approving over a mismatch.
const documentMaxRunes = 500

func refuse(code string) error { return &CodedError{Code: code} }

func (a Application) guardDecidable(operatorAccountID string) error {
	if a.AccountID == operatorAccountID {
		return refuse(CodeOperatorOwnApplication)
	}
	if a.State != StatePending {
		return refuse(CodeApplicationNotPending)
	}
	return nil
}

// expectedEDRPOU is the code the extract must show: a transfer is checked
// against the association's code when known, otherwise the filed one.
func (a Application) expectedEDRPOU(associationEDRPOU string) string {
	if a.Kind == KindTransfer && associationEDRPOU != "" {
		return associationEDRPOU
	}
	return a.EDRPOU
}

func validMismatchDocument(doc string) bool {
	return strings.TrimSpace(doc) != "" && utf8.RuneCountInString(doc) <= documentMaxRunes
}

// RecordCheck stores the оператор's finding; it is not guarded.
func (a Application) RecordCheck(c Check) Application {
	a.Check = &c
	return a
}

// Approve decides the application in favour of the applicant.
func (a Application) Approve(in ApprovalInput) (Application, error) {
	if err := a.guardDecidable(in.OperatorAccountID); err != nil {
		return a, err
	}
	if a.Check == nil || !a.Check.SignatureValid {
		return a, refuse(CodeSignatureNotChecked)
	}
	if a.Check.ExtractEDRPOU != a.expectedEDRPOU(in.AssociationEDRPOU) {
		return a, refuse(CodeExtractCodeMismatch)
	}
	if in.MismatchOpen && !validMismatchDocument(in.Document) {
		return a, refuse(CodeMismatchNeedsDocument)
	}
	a.State = StateApproved
	return a, nil
}

// Reject decides the application against the applicant, with a reason.
func (a Application) Reject(operatorAccountID, reason string) (Application, error) {
	if err := a.guardDecidable(operatorAccountID); err != nil {
		return a, err
	}
	if strings.TrimSpace(reason) == "" {
		return a, refuse(CodeReasonRequired)
	}
	a.State = StateRejected
	return a, nil
}

// Refile reopens a rejected application for a fresh check, by its applicant only.
func (a Application) Refile(applicantAccountID string) (Application, error) {
	if a.State != StateRejected || a.AccountID != applicantAccountID {
		return a, refuse(CodeApplicationNotPending)
	}
	a.State = StatePending
	a.Check = nil
	return a, nil
}
