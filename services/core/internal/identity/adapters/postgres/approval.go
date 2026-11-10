package postgres

import (
	"context"
	"errors"

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

func (s *Approvals) Approve(ctx context.Context, req ports.ApprovalRequest, decide func(ports.ApprovalFacts) error, seat ports.SeatHeadFunc) (ports.ReviewDecision, error) {
	assocID, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	decisionID, err := uuid.NewV7()
	if err != nil {
		return ports.ReviewDecision{}, err
	}
	var out ports.ReviewDecision
	// The scope is the association being created; the actor is set to the applicant once known.
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
		if app.Kind != "connection" {
			return ports.ErrTransferNotSupported
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
		if err := decide(facts); err != nil {
			return err
		}

		if err := q.CreateAssociation(ctx, sqlcgen.CreateAssociationParams{
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
