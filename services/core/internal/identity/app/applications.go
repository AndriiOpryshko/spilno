package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Applications files заявки and keeps their extracts.
type Applications struct {
	repo   ports.ApplicationRepository
	roster ports.Roster
	now    func() time.Time
}

func NewApplications(repo ports.ApplicationRepository) *Applications {
	return &Applications{repo: repo, now: time.Now}
}

// UseRoster sets the roster that vouches for an association_id sent with a запит на передачу.
func (a *Applications) UseRoster(r ports.Roster) { a.roster = r }

// FileInput is a заявка as the person typed it.
type FileInput struct {
	AccountID     uuid.UUID
	Email         string
	AssociationID *uuid.UUID // only for a transfer filed from the card (AC-19c)
	Filing        domain.Filing
}

// File validates the filing, then records it pending. A refusal is a *domain.CodedFieldsError or a
// *domain.CodedError.
func (a *Applications) File(ctx context.Context, in FileInput) (ports.FiledApplication, error) {
	f := in.Filing
	if err := f.Validate(a.now()); err != nil {
		return ports.FiledApplication{}, err
	}
	nf := ports.NewFiling{
		AccountID: in.AccountID, Kind: string(f.Kind), Edrpou: f.Edrpou,
		AssociationName: f.Name, AssociationAddress: f.Address, ApplicantFullName: f.ApplicantName,
		Position: f.Position, TermStart: f.TermStart, TermEnd: f.TermEnd,
		ExtractName: f.Extract.Name, ExtractMediaType: http.DetectContentType(f.Extract.Content),
		ExtractContent: f.Extract.Content,
	}
	var out ports.FiledApplication
	var err error
	if f.Kind == domain.FilingTransfer {
		out, err = a.repo.FileTransfer(ctx, ports.NewTransfer{NewFiling: nf, ApplicantEmail: in.Email, AssociationID: in.AssociationID, Roster: a.roster})
	} else {
		out, err = a.repo.File(ctx, nf)
	}
	switch err {
	case ports.ErrAssociationNotConnected:
		return out, &domain.CodedError{Code: domain.CodeAssociationNotConnected}
	case ports.ErrAssociationConnected:
		return out, &domain.CodedError{Code: domain.CodeAssociationAlreadyConnected}
	case ports.ErrApplicationOpen:
		return out, &domain.CodedError{Code: domain.CodeApplicationAlreadyOpen}
	}
	return out, err
}

// ErrApplicationNotFound is the one answer for a missing id and another account's id.
var ErrApplicationNotFound = ports.ErrApplicationNotFound

// ErrExtractCorrupted is returned when a stored extract no longer matches its stored hash.
var ErrExtractCorrupted = errors.New("identity: stored extract does not match its sha256")

// ListMine returns the account's applications, newest first.
func (a *Applications) ListMine(ctx context.Context, account uuid.UUID, after *ports.ApplicationCursor, limit int) ([]ports.ApplicationView, error) {
	return a.repo.ListOfAccount(ctx, account, after, limit)
}

// GetMine returns ErrApplicationNotFound unless the application is the account's.
func (a *Applications) GetMine(ctx context.Context, account, id uuid.UUID) (ports.ApplicationDetail, error) {
	return a.repo.GetOfAccount(ctx, account, id)
}

// RefileInput is the new filing of a rejected application; the code is the application's own.
type RefileInput struct {
	AccountID     uuid.UUID
	ApplicationID uuid.UUID
	Filing        domain.Filing
}

// Refile reopens a rejected application with a new filing and extract.
func (a *Applications) Refile(ctx context.Context, in RefileInput) (ports.ApplicationView, error) {
	out, err := a.repo.Refile(ctx, in.AccountID, in.ApplicationID, func(s ports.StoredApplication) (ports.NewFiling, error) {
		cur := domain.Application{
			Kind: domain.ApplicationKind(s.Kind), State: domain.ApplicationState(s.State),
			AccountID: in.AccountID.String(), EDRPOU: s.Edrpou,
		}
		if _, err := cur.Refile(in.AccountID.String()); err != nil {
			return ports.NewFiling{}, err
		}
		f := in.Filing
		f.Kind, f.Edrpou = domain.FilingKind(s.Kind), s.Edrpou
		if err := f.Validate(a.now()); err != nil {
			return ports.NewFiling{}, err
		}
		return ports.NewFiling{
			AccountID: in.AccountID, Kind: s.Kind, Edrpou: s.Edrpou,
			AssociationName: f.Name, AssociationAddress: f.Address, ApplicantFullName: f.ApplicantName,
			Position: f.Position, TermStart: f.TermStart, TermEnd: f.TermEnd,
			ExtractName: f.Extract.Name, ExtractMediaType: http.DetectContentType(f.Extract.Content),
			ExtractContent: f.Extract.Content,
		}, nil
	})
	switch err {
	case ports.ErrAssociationConnected:
		return out, &domain.CodedError{Code: domain.CodeAssociationAlreadyConnected}
	case ports.ErrApplicationOpen:
		return out, &domain.CodedError{Code: domain.CodeApplicationAlreadyOpen}
	}
	return out, err
}

// Extract returns the account's own extract after comparing its hash with the stored one.
func (a *Applications) Extract(ctx context.Context, account, application, filing uuid.UUID) (ports.StoredExtract, error) {
	e, err := a.repo.GetExtract(ctx, account, application, filing)
	if err != nil {
		return e, err
	}
	if sum := sha256.Sum256(e.Content); !bytes.Equal(sum[:], e.Sha256) {
		return ports.StoredExtract{}, ErrExtractCorrupted
	}
	return e, nil
}
