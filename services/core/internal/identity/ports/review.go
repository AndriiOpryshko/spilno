package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrApplicationNotFound is returned when no application (or filing of it) has the id.
var ErrApplicationNotFound = errors.New("identity: application not found")

// ReviewApplication is an application with its latest filing's name and date.
type ReviewApplication struct {
	ID              uuid.UUID
	Kind            string
	Edrpou          string
	AssociationID   *uuid.UUID
	State           string
	CreatedAt       time.Time
	AssociationName string
	FiledAt         time.Time
	RejectionReason *string
}

// ReviewFiling is one filing with its extract's metadata.
type ReviewFiling struct {
	ID                 uuid.UUID
	AssociationName    string
	AssociationAddress string
	ApplicantFullName  string
	Position           string
	TermStart          time.Time
	TermEnd            *time.Time
	CreatedAt          time.Time
	ExtractID          uuid.UUID
	ExtractMediaType   string
	ExtractName        string
	ExtractSha256      []byte
	ExtractSizeBytes   int64
}

// ReviewCheck is a recorded check of an extract.
type ReviewCheck struct {
	ID             uuid.UUID
	FilingID       uuid.UUID
	SignatureValid bool
	SignerFullName string
	SignedOn       time.Time
	HeadFullName   string
	ExtractEdrpou  string
	ExtractDatedOn time.Time
	CheckedAt      time.Time
}

// ReviewDecision is a recorded decision.
type ReviewDecision struct {
	ID               uuid.UUID
	FilingID         uuid.UUID
	AssociationID    *uuid.UUID
	Outcome          string
	Reason           *string
	ReliedOnDocument *string
	DecidedAt        time.Time
}

// ReviewDetail is everything the оператор sees of an application. Filings are newest first,
// checks and competing applications oldest first.
type ReviewDetail struct {
	Application    ReviewApplication
	ApplicantEmail string
	Filings        []ReviewFiling
	Checks         []ReviewCheck
	Decisions      []ReviewDecision
	Competing      []uuid.UUID
}

// NewCheck is a check the оператор records.
type NewCheck struct {
	ApplicationID  uuid.UUID
	FilingID       uuid.UUID
	SignatureValid bool
	SignerFullName string
	SignedOn       time.Time
	HeadFullName   string
	ExtractEdrpou  string
	ExtractDatedOn time.Time
	CheckedByEmail string
}

// ReviewExtract is a stored extract as uploaded.
type ReviewExtract struct {
	Content      []byte
	MediaType    string
	OriginalName string
}

// ReviewRepository reads and writes what the оператор reviews.
type ReviewRepository interface {
	// List pages applications in a state: pending oldest first, the others newest first.
	List(ctx context.Context, state string, after *AssociationCursor, limit int) ([]ReviewApplication, error)
	// Get returns ErrApplicationNotFound for an unknown id.
	Get(ctx context.Context, id uuid.UUID) (ReviewDetail, error)
	// Extract returns ErrApplicationNotFound when the filing is not of the application.
	Extract(ctx context.Context, applicationID, filingID uuid.UUID) (ReviewExtract, error)
	// RecordCheck inserts a new row; ErrApplicationNotFound when the filing is not of the application.
	RecordCheck(ctx context.Context, c NewCheck) (ReviewCheck, error)
}
