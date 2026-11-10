package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Approvals implements ports.ApprovalRepository.
type Approvals struct {
	pool *pgxpool.Pool
}

var _ ports.ApprovalRepository = (*Approvals)(nil)

func NewApprovals(pool *pgxpool.Pool) *Approvals { return &Approvals{pool: pool} }

func (s *Approvals) Approve(ctx context.Context, req ports.ApprovalRequest, decide func(ports.ApprovalFacts) error, seat ports.SeatHeadFunc, end ports.EndTermFunc) (ports.ReviewDecision, error) {
	assocID, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	decisionID, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	scope, err := sqlcgen.New(s.pool).GetApplicationScope(ctx, pgUUID(req.ApplicationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ReviewDecision{}, ports.ErrApplicationNotFound
	}
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	transfer := scope.Kind != "connection"
	if transfer {
		assocID = scope.AssociationID.Bytes
	}
	var out ports.ReviewDecision
	// The scope is the association being created, or the one a transfer is for.
	err = db.InAssociation(ctx, s.pool, assocID, req.OperatorEmail, func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		q := sqlcgen.New(tx)
		app, err := q.GetApplicationForApproval(ctx, pgUUID(req.ApplicationID))
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrApplicationNotFound
		}
		if err != nil {
			return err
		}
		filing, err := q.GetLatestFilingForApproval(ctx, pgUUID(req.ApplicationID))
		if err != nil {
			return err
		}
		facts := ports.ApprovalFacts{
			Kind: app.Kind, State: app.State, AccountID: app.AccountID.Bytes, Edrpou: app.Edrpou,
			FilingIsLatest: filing.ID.Bytes == req.FilingID, ApplicantName: filing.ApplicantFullName,
		}
		var checkID pgtype.UUID
		check, err := q.GetLatestCheck(ctx, filing.ID)
		switch {
		case err == nil:
			checkID = check.ID
			facts.HasCheck, facts.SignatureValid, facts.ExtractEdrpou = true, check.SignatureValid, check.ExtractEdrpou
			facts.SignerFullName, facts.HeadFullName, facts.ExtractDatedOn = check.SignerFullName, check.HeadFullName, check.ExtractDatedOn.Time
		case !errors.Is(err, pgx.ErrNoRows):
			return err
		}
		facts.FiledAt = filing.CreatedAt.Time
		var assoc sqlcgen.GetAssociationForApprovalRow
		if transfer {
			if assoc, err = q.GetAssociationForApproval(ctx, pgUUID(assocID)); err != nil {
				return err
			}
			facts.AssociationEdrpou = assoc.Edrpou.String
		}
		if err := decide(facts); err != nil {
			return err
		}

		if transfer {
			if err := q.SetAssociationEdrpou(ctx, sqlcgen.SetAssociationEdrpouParams{ID: pgUUID(assocID), Edrpou: pgtype.Text{String: app.Edrpou, Valid: true}}); err != nil {
				return err
			}
			if err := endPreviousHead(ctx, q, end, assocID, assoc.ConfirmedByDecisionID, app.ApplicantEmail, filing.TermStart.Time, decisionID); err != nil {
				return err
			}
		} else if err := q.CreateAssociation(ctx, sqlcgen.CreateAssociationParams{
			ID: pgUUID(assocID), Name: filing.AssociationName, Edrpou: pgtype.Text{String: app.Edrpou, Valid: true},
			Address: pgtype.Text{String: filing.AssociationAddress, Valid: true},
		}); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) && pg.Code == "23505" {
				return ports.ErrAssociationConnected
			}
			return err
		}
		if _, err := seat(ctx, assocID, filing.ApplicantFullName, app.ApplicantEmail, filing.TermStart.Time, decisionID); err != nil {
			return err
		}
		doc := pgtype.Text{String: req.Document, Valid: req.Document != ""}
		at, err := q.CreateApprovalDecision(ctx, sqlcgen.CreateApprovalDecisionParams{
			ID: pgUUID(decisionID), ApplicationID: pgUUID(req.ApplicationID), FilingID: filing.ID, CheckID: checkID,
			AssociationID: pgUUID(assocID), ReliedOnDocument: doc, DecidedByEmail: req.OperatorEmail,
		})
		if err != nil {
			return err
		}
		if err := q.SetApplicationApproved(ctx, sqlcgen.SetApplicationApprovedParams{ID: pgUUID(req.ApplicationID), AssociationID: pgUUID(assocID)}); err != nil {
			return err
		}
		if err := q.SetAssociationConfirmation(ctx, sqlcgen.SetAssociationConfirmationParams{ID: pgUUID(assocID), DecisionID: pgUUID(decisionID)}); err != nil {
			return err
		}
		out = ports.ReviewDecision{
			ID: decisionID, FilingID: filing.ID.Bytes, AssociationID: &assocID, Outcome: "approved",
			ReliedOnDocument: optText(doc), DecidedAt: at.Time,
		}
		return nil
	})
	return out, err
}

// endPreviousHead ends the whole entry of the голова confirmed by the earlier approval the day before
// the new term starts, never before the entry's own first day. Nothing ends when no approval
// confirmed a голова or when it is the applicant, who is re-seated.
func endPreviousHead(ctx context.Context, q *sqlcgen.Queries, end ports.EndTermFunc, assocID uuid.UUID, confirmedBy pgtype.UUID, applicantEmail string, termStart time.Time, decisionID uuid.UUID) error {
	if !confirmedBy.Valid {
		return nil
	}
	prev, err := q.GetEntryOfDecision(ctx, sqlcgen.GetEntryOfDecisionParams{AssociationID: pgUUID(assocID), DecisionID: confirmedBy})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.EqualFold(prev.SignInEmail, applicantEmail) {
		return nil
	}
	date := termStart.AddDate(0, 0, -1)
	if date.Before(prev.TermStart.Time) {
		date = prev.TermStart.Time
	}
	return end(ctx, assocID, prev.BoardMemberID.Bytes, date, decisionID)
}
