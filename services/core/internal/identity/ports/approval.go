package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrTransferNotSupported is returned when the application is a запит на передачу, approved elsewhere.
var ErrTransferNotSupported = errors.New("identity: approving a transfer is not supported here")

// SeatHeadFunc is the roster step run inside the approval's transaction.
type SeatHeadFunc func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error)

// ApprovalRequest is what the оператор sends to approve.
type ApprovalRequest struct {
	ApplicationID uuid.UUID
	FilingID      uuid.UUID
	OperatorEmail string
	Document      string
}

// ApprovalFacts is what the guards of an approval look at, read under the application's lock.
type ApprovalFacts struct {
	Kind           string
	State          string
	AccountID      uuid.UUID
	Edrpou         string
	FilingIsLatest bool
	ApplicantName  string
	FiledAt        time.Time
	HasCheck       bool
	SignatureValid bool
	ExtractEdrpou  string
	SignerFullName string
	HeadFullName   string
	ExtractDatedOn time.Time
}

// ApprovalRepository applies an approval atomically.
type ApprovalRepository interface {
	// Approve reads the facts, runs decide on them, then creates the association, seats the head
	// through seat, records the decision and approves the application, in one transaction.
	// ErrApplicationNotFound for an unknown id; ErrAssociationConnected when the code was
	// connected meanwhile; ErrTransferNotSupported for a transfer.
	Approve(ctx context.Context, req ApprovalRequest, decide func(ApprovalFacts) error, seat SeatHeadFunc) (ReviewDecision, error)
}
