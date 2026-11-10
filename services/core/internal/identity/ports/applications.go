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
}
