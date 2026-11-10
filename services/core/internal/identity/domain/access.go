package domain

// PathKind says which grant a request asks for.
type PathKind int

const (
	AssociationPath PathKind = iota
	OperatorPath
)

// Actor is the signed-in person as the identity middleware knows them.
type Actor struct {
	AccountID      string
	Email          string
	AssociationID  string
	Member         bool   // a current roster entry in AssociationID
	ReadOnly       bool   // AssociationID is disabled
	DisabledReason string // the reason shown when ReadOnly refuses
	Operator       bool
}

// Decision is the verdict for one request.
type Decision struct {
	Allowed bool
	Code    string
	Reason  string
}

// Decide is the pure access decision per request.
func Decide(a Actor, path PathKind, operationID string) Decision {
	if path == OperatorPath {
		if a.Operator {
			return Decision{Allowed: true}
		}
		return Decision{Code: CodeNotOperator}
	}
	if !a.Member {
		return Decision{Code: CodeNotCurrentMember}
	}
	if a.ReadOnly {
		switch operationID {
		case "searchRegistry", "getSignedProtocol", "downloadFile":
		default:
			return Decision{Code: CodeAssociationDisabled, Reason: a.DisabledReason}
		}
	}
	return Decision{Allowed: true}
}
