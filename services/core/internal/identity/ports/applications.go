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
}
