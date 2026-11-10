package app

import (
	"context"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
)

// Approvals approves a заявка на підключення.
type Approvals struct {
	repo ports.ApprovalRepository
	seat ports.SeatHeadFunc
	end  ports.EndTermFunc
}

func NewApprovals(repo ports.ApprovalRepository, seat ports.SeatHeadFunc, end ports.EndTermFunc) *Approvals {
	return &Approvals{repo: repo, seat: seat, end: end}
}

// ApproveInput is an approval by the оператор.
type ApproveInput struct {
	ports.ApprovalRequest
	OperatorAccountID string
}

// Approve runs the domain guards on the facts read under lock, then applies the approval.
func (a *Approvals) Approve(ctx context.Context, in ApproveInput) (ports.ReviewDecision, error) {
	return a.repo.Approve(ctx, in.ApprovalRequest, func(f ports.ApprovalFacts) error {
		if !f.FilingIsLatest {
			return &domain.CodedError{Code: domain.CodeApplicationNotPending}
		}
		app := domain.Application{
			Kind: domain.ApplicationKind(f.Kind), State: domain.ApplicationState(f.State),
			AccountID: f.AccountID.String(), EDRPOU: f.Edrpou,
		}
		mismatch := false
		if f.HasCheck {
			app = app.RecordCheck(domain.Check{SignatureValid: f.SignatureValid, ExtractEDRPOU: f.ExtractEdrpou})
			mismatch = nameMismatchOpen(f)
		}
		_, err := app.Approve(domain.ApprovalInput{
			OperatorAccountID: in.OperatorAccountID, AssociationEDRPOU: f.AssociationEdrpou, MismatchOpen: mismatch, Document: in.Document,
		})
		return err
	}, a.seat, a.end)
}

func nameMismatchOpen(f ports.ApprovalFacts) bool {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		loc = time.UTC
	}
	y, m, d := f.FiledAt.In(loc).Date()
	for _, w := range domain.ReviewWarnings(f.ApplicantName, domain.ExtractCheck{
		SignerName: f.SignerFullName, HeadFullName: f.HeadFullName, ExtractDated: f.ExtractDatedOn,
	}, time.Date(y, m, d, 0, 0, 0, 0, time.UTC)) {
		if w.Kind == domain.WarningNameMismatch {
			return true
		}
	}
	return false
}
