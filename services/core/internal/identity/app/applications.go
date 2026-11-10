package app

import (
	"context"
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
