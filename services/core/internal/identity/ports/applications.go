package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrAssociationConnected is returned when the code already belongs to a connected association.
	ErrAssociationConnected = errors.New("identity: association already connected")
	// ErrApplicationOpen is returned when the account already has a pending заявка for the code.
	ErrApplicationOpen = errors.New("identity: application already open")
	// ErrAssociationNotConnected is returned when a запит на передачу finds no association to attach to.
	ErrAssociationNotConnected = errors.New("identity: association not connected")
	// ErrApplicationNotFound is returned when the account has no such application, filing or extract.
	ErrApplicationNotFound = errors.New("identity: application not found")
)

// NewFiling is one first filing of a new application with its extract, as uploaded.
type NewFiling struct {
	AccountID          uuid.UUID
	Kind               string
	Edrpou             string
	AssociationName    string
	AssociationAddress string
	ApplicantFullName  string
	Position           string
	TermStart          time.Time
	TermEnd            *time.Time
	ExtractName        string
	ExtractMediaType   string
	ExtractContent     []byte
}

// Roster tells whether an email is a current член правління of an association, on the transaction in ctx.
type Roster interface {
	CurrentMember(ctx context.Context, association uuid.UUID, email string, date time.Time) (bool, error)
}

// NewTransfer is a запит на передачу: the filing plus how it finds its association.
type NewTransfer struct {
	NewFiling
	ApplicantEmail string
	// AssociationID is set when filed from the card of an association set up by hand.
	AssociationID *uuid.UUID
	Roster        Roster
}

// FiledApplication is what filing recorded.
type FiledApplication struct {
	ID        uuid.UUID
	CreatedAt time.Time
	FiledAt   time.Time
}

// ApplicationRepository stores applications, their filings and extracts.
type ApplicationRepository interface {
	// File refuses a connected code (ErrAssociationConnected) and an open заявка of the account for
	// the code (ErrApplicationOpen), otherwise stores the extract, the pending application and its
	// filing in one transaction.
	File(ctx context.Context, f NewFiling) (FiledApplication, error)
	// FileTransfer attaches the запит to the association found by AssociationID (the applicant must be
	// its current member, checked inside db.InAssociation) or else by the code, connected or disabled;
	// none found is ErrAssociationNotConnected. Otherwise it stores like File.
	FileTransfer(ctx context.Context, f NewTransfer) (FiledApplication, error)
	// ListOfAccount returns the account's applications newest first, after the cursor.
	ListOfAccount(ctx context.Context, account uuid.UUID, after *ApplicationCursor, limit int) ([]ApplicationView, error)
	// GetOfAccount returns ErrApplicationNotFound for a missing or another account's application.
	GetOfAccount(ctx context.Context, account, id uuid.UUID) (ApplicationDetail, error)
	// Refile locks the account's application (ErrApplicationNotFound otherwise), lets prepare judge
	// it and build the filing, refuses a connection whose code was connected meanwhile
	// (ErrAssociationConnected), then stores the extract and filing and sets the state pending.
	Refile(ctx context.Context, account, id uuid.UUID, prepare func(StoredApplication) (NewFiling, error)) (ApplicationView, error)
	// GetExtract returns ErrApplicationNotFound unless the filing belongs to the account's application.
	GetExtract(ctx context.Context, account, application, filing uuid.UUID) (StoredExtract, error)
}

// ApplicationCursor is the keyset position over (created_at, id), newest first.
type ApplicationCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// StoredApplication is the row a refile is judged on.
type StoredApplication struct {
	Kind   string
	State  string
	Edrpou string
}

// ApplicationView is an application with its latest filing's name and date.
type ApplicationView struct {
	ID              uuid.UUID
	Kind            string
	State           string
	Edrpou          string
	AssociationID   *uuid.UUID
	AssociationName string
	CreatedAt       time.Time
	FiledAt         time.Time
	RejectionReason *string
}

// FilingView is one filing with its extract's metadata.
type FilingView struct {
	ID                 uuid.UUID
	AssociationName    string
	AssociationAddress string
	ApplicantFullName  string
	Position           string
	TermStart          time.Time
	TermEnd            *time.Time
	CreatedAt          time.Time
	ExtractID          uuid.UUID
	MediaType          string
	OriginalName       string
	Sha256             []byte
	SizeBytes          int64
}

// DecisionView is the оператор's decision on one filing.
type DecisionView struct {
	ID               uuid.UUID
	FilingID         uuid.UUID
	AssociationID    *uuid.UUID
	Outcome          string
	Reason           *string
	ReliedOnDocument *string
	CreatedAt        time.Time
}

// ApplicationDetail is the applicant's view; filings and decisions newest first.
type ApplicationDetail struct {
	Application ApplicationView
	Filings     []FilingView
	Decisions   []DecisionView
}

// StoredExtract is an uploaded file with the hash stored at upload.
type StoredExtract struct {
	Name      string
	MediaType string
	Content   []byte
	Sha256    []byte
}
