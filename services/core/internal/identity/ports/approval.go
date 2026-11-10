package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SeatHeadFunc is the roster step run inside the approval's transaction.
type SeatHeadFunc func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error)

// EndTermFunc ends the whole roster entry memberID on date, inside the approval's transaction.
type EndTermFunc func(ctx context.Context, association, memberID uuid.UUID, date time.Time, decisionID uuid.UUID) error

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
	// AssociationEdrpou is the code of the association a transfer is for, empty when it has none.
	AssociationEdrpou string
}

// ApprovalRepository applies an approval atomically.
type ApprovalRepository interface {
	// Approve reads the facts, runs decide on them, then creates the association, seats the head
	// through seat, records the decision and approves the application, in one transaction.
	// ErrApplicationNotFound for an unknown id; ErrAssociationConnected when the code was
	// connected meanwhile. A transfer runs in the association of the запит: it ends the entry of the
	// голова confirmed by the earlier approval through end, then seats the applicant.
	Approve(ctx context.Context, req ApprovalRequest, decide func(ApprovalFacts) error, seat SeatHeadFunc, end EndTermFunc) (ReviewDecision, error)
}
