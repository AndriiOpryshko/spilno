package domain

import "testing"

// Operation ids of the live contract and of the feature delta.
const (
	opSearchRegistry     = "searchRegistry"
	opGetSignedProtocol  = "getSignedProtocol"
	opDownloadFile       = "downloadFile"
	opEditDraft          = "updateProtocolDraft"
	opUploadDocument     = "uploadDocument"
	opStartStorage       = "startStorageConnection"
	opGetAssociationCard = "getAssociationCard"
	opNewUnlistedOp      = "someFutureOperation"

	opListForReview = "listApplicationsForReview"
	opApprove       = "approveApplication"
	opReject        = "rejectApplication"
	opDisable       = "disableAssociation"
	opEnable        = "enableAssociation"
)

var (
	outsider       = Actor{AccountID: "a1", Email: "p@example.test", AssociationID: "A"}
	connected      = Actor{AccountID: "a2", Email: "m@example.test", AssociationID: "A", Member: true}
	disabledMember = Actor{AccountID: "a3", Email: "d@example.test", AssociationID: "A", Member: true, ReadOnly: true, DisabledReason: "Подвійне підключення"}
	operatorOnly   = Actor{AccountID: "a4", Email: "o@example.test", AssociationID: "A", Operator: true}
	operatorMember = Actor{AccountID: "a5", Email: "om@example.test", AssociationID: "A", Member: true, Operator: true}
	operatorMemOfA = Actor{AccountID: "a5", Email: "om@example.test", AssociationID: "B", Member: false, Operator: true}
)

func TestDecide_PRDSection10Matrix(t *testing.T) {
	read := []string{opGetAssociationCard, opSearchRegistry, opGetSignedProtocol, opDownloadFile}
	change := []string{opEditDraft, opUploadDocument}
	storage := []string{opStartStorage}

	type cell struct {
		name string
		a    Actor
		path PathKind
		ops  []string
		code string // "" = allowed
	}
	all := func(groups ...[]string) []string {
		var o []string
		for _, g := range groups {
			o = append(o, g...)
		}
		return o
	}
	cells := []cell{
		// голова, off the roster (AC-07)
		{"applicant reads association", outsider, AssociationPath, read, CodeNotCurrentMember},
		{"applicant changes association", outsider, AssociationPath, change, CodeNotCurrentMember},
		{"applicant connects storage", outsider, AssociationPath, storage, CodeNotCurrentMember},
		{"applicant decides", outsider, OperatorPath, []string{opApprove, opReject, opListForReview}, CodeNotOperator},
		{"applicant disables", outsider, OperatorPath, []string{opDisable, opEnable}, CodeNotOperator},

		// член правління of a connected association
		{"member reads", connected, AssociationPath, read, ""},
		{"member changes", connected, AssociationPath, change, ""},
		{"member connects storage", connected, AssociationPath, storage, ""},
		{"member decides", connected, OperatorPath, []string{opApprove, opReject, opListForReview}, CodeNotOperator},
		{"member disables", connected, OperatorPath, []string{opDisable, opEnable}, CodeNotOperator},

		// член правління of a disabled association (AC-25, AC-26)
		{"disabled member reads registry", disabledMember, AssociationPath, []string{opSearchRegistry, opGetSignedProtocol, opDownloadFile}, ""},
		{"disabled member changes", disabledMember, AssociationPath, change, CodeAssociationDisabled},
		{"disabled member connects storage", disabledMember, AssociationPath, storage, CodeAssociationDisabled},
		{"disabled member reads card", disabledMember, AssociationPath, []string{opGetAssociationCard}, CodeAssociationDisabled},
		{"disabled member decides", disabledMember, OperatorPath, []string{opApprove, opDisable}, CodeNotOperator},

		// оператор only (AC-16b)
		{"operator decides", operatorOnly, OperatorPath, all([]string{opApprove, opReject, opListForReview, opDisable, opEnable}), ""},
		{"operator reads association records", operatorOnly, AssociationPath, read, CodeNotCurrentMember},
		{"operator changes association records", operatorOnly, AssociationPath, change, CodeNotCurrentMember},
		{"operator connects storage", operatorOnly, AssociationPath, storage, CodeNotCurrentMember},

		// оператор who is also a member of A: neither role widens the other
		{"operator-member decides", operatorMember, OperatorPath, []string{opApprove, opDisable}, ""},
		{"operator-member reads own", operatorMember, AssociationPath, read, ""},
		{"operator-member changes own", operatorMember, AssociationPath, change, ""},
		// and the operator role opens nothing in an association they are not on
		{"operator-member reads other", operatorMemOfA, AssociationPath, read, CodeNotCurrentMember},
		{"operator-member changes other", operatorMemOfA, AssociationPath, change, CodeNotCurrentMember},
	}

	for _, c := range cells {
		for _, op := range c.ops {
			t.Run(c.name+"/"+op, func(t *testing.T) {
				d := Decide(c.a, c.path, op)
				if c.code == "" {
					if !d.Allowed || d.Code != "" {
						t.Fatalf("Decide(%s) = %+v, want allowed", op, d)
					}
					return
				}
				if d.Allowed || d.Code != c.code {
					t.Fatalf("Decide(%s) = %+v, want refused with %q", op, d, c.code)
				}
			})
		}
	}
}

func TestDecide_DisabledRefusalCarriesReason(t *testing.T) {
	d := Decide(disabledMember, AssociationPath, opEditDraft)
	if d.Code != CodeAssociationDisabled || d.Reason != "Подвійне підключення" {
		t.Fatalf("Decide = %+v, want association_disabled with the reason", d)
	}
}

func TestDecide_DisabledIsFailClosedForUnlistedOperation(t *testing.T) {
	d := Decide(disabledMember, AssociationPath, opNewUnlistedOp)
	if d.Allowed || d.Code != CodeAssociationDisabled {
		t.Fatalf("Decide = %+v, want refused: an operation off the allow-list", d)
	}
	if d := Decide(connected, AssociationPath, opNewUnlistedOp); !d.Allowed {
		t.Fatalf("a connected member must pass an unlisted operation, got %+v", d)
	}
}

func TestDecide_NonMemberIsRefusedWhateverTheState(t *testing.T) {
	// no member is not_current_member, whatever the state and the operator flag;
	// the read-only flag must not turn it into association_disabled.
	for _, a := range []Actor{
		{AssociationID: "A", ReadOnly: true, DisabledReason: "x"},
		{AssociationID: "A", ReadOnly: true, DisabledReason: "x", Operator: true},
		{AssociationID: "A", Operator: true},
	} {
		for _, op := range []string{opSearchRegistry, opEditDraft} {
			d := Decide(a, AssociationPath, op)
			if d.Allowed || d.Code != CodeNotCurrentMember {
				t.Errorf("Decide(%+v, %s) = %+v, want not_current_member", a, op, d)
			}
		}
	}
}

func TestDecide_OperatorPathAsksTheOperatorFlagOnly(t *testing.T) {
	// a member of a disabled association keeps the operator role's rights.
	a := Actor{AssociationID: "A", Member: true, ReadOnly: true, DisabledReason: "x", Operator: true}
	if d := Decide(a, OperatorPath, opApprove); !d.Allowed {
		t.Fatalf("Decide = %+v, want allowed on the operator path", d)
	}
	// no roster membership is needed on an operator path.
	if d := Decide(Actor{Operator: true}, OperatorPath, opDisable); !d.Allowed {
		t.Fatalf("Decide = %+v, want allowed", d)
	}
}
