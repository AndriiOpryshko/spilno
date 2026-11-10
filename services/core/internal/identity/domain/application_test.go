package domain

import (
	"errors"
	"strings"
	"testing"
)

const (
	acctOperator  = "account-operator"
	acctApplicant = "account-applicant"
)

func okCheck() *Check { return &Check{SignatureValid: true, ExtractEDRPOU: "12345678"} }

func app(kind ApplicationKind, state ApplicationState, filer string, check *Check) Application {
	return Application{Kind: kind, State: state, AccountID: filer, EDRPOU: "12345678", Check: check}
}

func wantCode(t *testing.T, err error, code string) {
	t.Helper()
	var ce *CodedError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want *CodedError with code %q", err, code)
	}
	if ce.Code != code {
		t.Errorf("code = %q, want %q", ce.Code, code)
	}
}

func TestApplicationApprove(t *testing.T) {
	kinds := []ApplicationKind{KindConnection, KindTransfer}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			tests := []struct {
				name string
				app  Application
				in   ApprovalInput
				code string // empty = allowed
			}{
				{"allowed", app(kind, StatePending, acctApplicant, okCheck()), ApprovalInput{OperatorAccountID: acctOperator}, ""},
				{"allowed with document over a mismatch", app(kind, StatePending, acctApplicant, okCheck()),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true, Document: "протокол загальних зборів № 3 від 01.03.2026"}, ""},
				{"allowed with a document of exactly 500 characters", app(kind, StatePending, acctApplicant, okCheck()),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true, Document: strings.Repeat("я", 500)}, ""},
				{"own filing", app(kind, StatePending, acctOperator, okCheck()), ApprovalInput{OperatorAccountID: acctOperator}, CodeOperatorOwnApplication},
				{"approved already", app(kind, StateApproved, acctApplicant, okCheck()), ApprovalInput{OperatorAccountID: acctOperator}, CodeApplicationNotPending},
				{"rejected, not refiled", app(kind, StateRejected, acctApplicant, okCheck()), ApprovalInput{OperatorAccountID: acctOperator}, CodeApplicationNotPending},
				{"no check recorded", app(kind, StatePending, acctApplicant, nil), ApprovalInput{OperatorAccountID: acctOperator}, CodeSignatureNotChecked},
				{"signature found invalid", app(kind, StatePending, acctApplicant, &Check{SignatureValid: false, ExtractEDRPOU: "12345678"}),
					ApprovalInput{OperatorAccountID: acctOperator}, CodeSignatureNotChecked},
				{"extract of another code", app(kind, StatePending, acctApplicant, &Check{SignatureValid: true, ExtractEDRPOU: "87654321"}),
					ApprovalInput{OperatorAccountID: acctOperator}, CodeExtractCodeMismatch},
				{"mismatch without document", app(kind, StatePending, acctApplicant, okCheck()),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true}, CodeMismatchNeedsDocument},
				{"mismatch with blank document", app(kind, StatePending, acctApplicant, okCheck()),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true, Document: "   "}, CodeMismatchNeedsDocument},
				{"mismatch with document over 500 characters", app(kind, StatePending, acctApplicant, okCheck()),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true, Document: strings.Repeat("я", 501)}, CodeMismatchNeedsDocument},
				// guard order: own, not pending, signature, code, document
				{"order: own before decided", app(kind, StateApproved, acctOperator, nil), ApprovalInput{OperatorAccountID: acctOperator}, CodeOperatorOwnApplication},
				{"order: decided before unchecked", app(kind, StateRejected, acctApplicant, nil), ApprovalInput{OperatorAccountID: acctOperator}, CodeApplicationNotPending},
				{"order: unchecked before code", app(kind, StatePending, acctApplicant, &Check{SignatureValid: false, ExtractEDRPOU: "87654321"}),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true}, CodeSignatureNotChecked},
				{"order: code before document", app(kind, StatePending, acctApplicant, &Check{SignatureValid: true, ExtractEDRPOU: "87654321"}),
					ApprovalInput{OperatorAccountID: acctOperator, MismatchOpen: true}, CodeExtractCodeMismatch},
			}
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					got, err := tc.app.Approve(tc.in)
					if tc.code == "" {
						if err != nil {
							t.Fatalf("Approve: unexpected error %v", err)
						}
						if got.State != StateApproved {
							t.Errorf("state = %q, want %q", got.State, StateApproved)
						}
						return
					}
					wantCode(t, err, tc.code)
					if got.State != tc.app.State {
						t.Errorf("state after refusal = %q, want unchanged %q", got.State, tc.app.State)
					}
				})
			}
		})
	}
}

func TestApplicationApproveExtractCodeSource(t *testing.T) {
	t.Run("transfer compares with the association's code", func(t *testing.T) {
		a := app(KindTransfer, StatePending, acctApplicant, &Check{SignatureValid: true, ExtractEDRPOU: "11111111"})
		a.EDRPOU = "11111111" // typed code differs from the association's
		_, err := a.Approve(ApprovalInput{OperatorAccountID: acctOperator, AssociationEDRPOU: "12345678"})
		wantCode(t, err, CodeExtractCodeMismatch)
	})
	t.Run("transfer to the association with the same code", func(t *testing.T) {
		a := app(KindTransfer, StatePending, acctApplicant, okCheck())
		if _, err := a.Approve(ApprovalInput{OperatorAccountID: acctOperator, AssociationEDRPOU: "12345678"}); err != nil {
			t.Fatalf("Approve: %v", err)
		}
	})
	t.Run("hand-made association without a code uses the typed one", func(t *testing.T) {
		a := app(KindTransfer, StatePending, acctApplicant, okCheck())
		if _, err := a.Approve(ApprovalInput{OperatorAccountID: acctOperator, AssociationEDRPOU: ""}); err != nil {
			t.Fatalf("Approve: %v", err)
		}
	})
	t.Run("hand-made association: extract differing from the typed code", func(t *testing.T) {
		a := app(KindTransfer, StatePending, acctApplicant, &Check{SignatureValid: true, ExtractEDRPOU: "87654321"})
		_, err := a.Approve(ApprovalInput{OperatorAccountID: acctOperator})
		wantCode(t, err, CodeExtractCodeMismatch)
	})
}

func TestApplicationReject(t *testing.T) {
	for _, kind := range []ApplicationKind{KindConnection, KindTransfer} {
		t.Run(string(kind), func(t *testing.T) {
			tests := []struct {
				name   string
				app    Application
				reason string
				code   string
			}{
				{"allowed", app(kind, StatePending, acctApplicant, nil), "у витягу інший керівник, документа про обрання не надано", ""},
				{"allowed after an invalid signature", app(kind, StatePending, acctApplicant, &Check{SignatureValid: false}), "підпис недійсний", ""},
				{"own filing", app(kind, StatePending, acctOperator, nil), "причина", CodeOperatorOwnApplication},
				{"approved already", app(kind, StateApproved, acctApplicant, nil), "причина", CodeApplicationNotPending},
				{"rejected already", app(kind, StateRejected, acctApplicant, nil), "причина", CodeApplicationNotPending},
				{"empty reason", app(kind, StatePending, acctApplicant, nil), "", CodeReasonRequired},
				{"blank reason", app(kind, StatePending, acctApplicant, nil), "  \t ", CodeReasonRequired},
				{"order: own before decided", app(kind, StateRejected, acctOperator, nil), "", CodeOperatorOwnApplication},
				{"order: decided before reason", app(kind, StateApproved, acctApplicant, nil), "", CodeApplicationNotPending},
			}
			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					got, err := tc.app.Reject(acctOperator, tc.reason)
					if tc.code == "" {
						if err != nil {
							t.Fatalf("Reject: unexpected error %v", err)
						}
						if got.State != StateRejected {
							t.Errorf("state = %q, want %q", got.State, StateRejected)
						}
						return
					}
					wantCode(t, err, tc.code)
					if got.State != tc.app.State {
						t.Errorf("state after refusal = %q, want unchanged %q", got.State, tc.app.State)
					}
				})
			}
		})
	}
}

func TestApplicationRecordCheckIsNotGuarded(t *testing.T) {
	for _, st := range []ApplicationState{StatePending, StateApproved, StateRejected} {
		a := app(KindConnection, st, acctOperator, nil)
		got := a.RecordCheck(Check{SignatureValid: true, ExtractEDRPOU: "12345678"})
		if got.Check == nil || got.State != st {
			t.Errorf("state %q: RecordCheck gave check=%v state=%q, want the check recorded and state unchanged", st, got.Check, got.State)
		}
	}
}

func TestApplicationRefile(t *testing.T) {
	for _, kind := range []ApplicationKind{KindConnection, KindTransfer} {
		t.Run(string(kind), func(t *testing.T) {
			t.Run("rejected, by the same applicant", func(t *testing.T) {
				got, err := app(kind, StateRejected, acctApplicant, okCheck()).Refile(acctApplicant)
				if err != nil {
					t.Fatalf("Refile: %v", err)
				}
				if got.State != StatePending {
					t.Errorf("state = %q, want %q", got.State, StatePending)
				}
				if got.Check != nil {
					t.Errorf("check carried over to the new filing: %+v", got.Check)
				}
				// the refiled application needs a fresh check
				_, err = got.Approve(ApprovalInput{OperatorAccountID: acctOperator})
				wantCode(t, err, CodeSignatureNotChecked)
			})
			refused := []struct {
				name  string
				state ApplicationState
				who   string
			}{
				{"rejected, by another account", StateRejected, "account-other"},
				{"pending", StatePending, acctApplicant},
				{"approved", StateApproved, acctApplicant},
			}
			for _, tc := range refused {
				t.Run(tc.name, func(t *testing.T) {
					a := app(kind, tc.state, acctApplicant, okCheck())
					got, err := a.Refile(tc.who)
					if err == nil {
						t.Fatal("Refile: want a refusal, got nil")
					}
					if got.State != tc.state {
						t.Errorf("state after refusal = %q, want unchanged %q", got.State, tc.state)
					}
				})
			}
		})
	}
}
