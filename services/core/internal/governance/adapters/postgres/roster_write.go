package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// WriteRevision appends a revision of a roster entry on the transaction in ctx. The entry
// is validated in Go first: a missing or unknown position never reaches the database.
func (Roster) WriteRevision(ctx context.Context, association, boardMember uuid.UUID, entry domain.BoardMemberEntry) error {
	if err := validateForWrite(entry); err != nil {
		return err
	}
	var decision pgtype.UUID
	if entry.DecisionID != "" {
		id, err := uuid.Parse(entry.DecisionID)
		if err != nil {
			return &domain.CodedFieldsError{Code: domain.CodeRosterEntryInvalid, Fields: []domain.FieldIssue{{Pointer: "/decision_id", Detail: "must be a uuid"}}}
		}
		decision = uuidOf(id)
	}
	tx, err := db.From(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("revision id: %w", err)
	}
	params := sqlcgen.InsertBoardMemberRevisionParams{
		ID: uuidOf(id), AssociationID: uuidOf(association), BoardMemberID: uuidOf(boardMember),
		FullName:    entry.FullName,
		SignInEmail: pgtype.Text{String: entry.SignInEmail, Valid: entry.SignInEmail != ""},
		TermStart:   dateOf(entry.Term.Start),
		Position:    pgtype.Text{String: string(entry.Position), Valid: true},
		DecisionID:  decision,
	}
	if !entry.Term.End.IsZero() {
		params.TermEnd = dateOf(entry.Term.End)
	}
	return sqlcgen.New(tx).InsertBoardMemberRevision(ctx, params)
}

// validateForWrite is Validate plus the rule that a stored revision always has a position.
func validateForWrite(entry domain.BoardMemberEntry) error {
	err := entry.Validate()
	if entry.Position != "" {
		return err
	}
	var issues []domain.FieldIssue
	var coded *domain.CodedFieldsError
	if errors.As(err, &coded) {
		issues = append(issues, coded.Fields...)
	}
	issues = append(issues, domain.FieldIssue{Pointer: "/position", Detail: "is required"})
	return &domain.CodedFieldsError{Code: domain.CodeRosterEntryInvalid, Fields: issues}
}
